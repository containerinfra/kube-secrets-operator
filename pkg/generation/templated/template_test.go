package templated

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		secretData  map[string][]byte
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:     "simple template with single value",
			template: "Hello {{.Ref.username}}",
			secretData: map[string][]byte{
				"username": []byte("john"),
			},
			expected:    "Hello john",
			expectError: false,
		},
		{
			name:     "template with multiple values",
			template: "User: {{.Ref.username}}, Pass: {{.Ref.password}}",
			secretData: map[string][]byte{
				"username": []byte("john"),
				"password": []byte("secret123"),
			},
			expected:    "User: john, Pass: secret123",
			expectError: false,
		},
		{
			name:     "template with conditionals",
			template: "{{if .Ref.enabled}}Service is enabled{{else}}Service is disabled{{end}}",
			secretData: map[string][]byte{
				"enabled": []byte("true"),
			},
			expected:    "Service is enabled",
			expectError: false,
		},
		{
			name:     "template with string concatenation",
			template: "{{.Ref.protocol}}://{{.Ref.host}}:{{.Ref.port}}/{{.Ref.path}}",
			secretData: map[string][]byte{
				"protocol": []byte("https"),
				"host":     []byte("example.com"),
				"port":     []byte("8443"),
				"path":     []byte("api/v1"),
			},
			expected:    "https://example.com:8443/api/v1",
			expectError: false,
		},
		{
			name:     "template with special characters",
			template: "Connection: {{.Ref.dsn}}",
			secretData: map[string][]byte{
				"dsn": []byte("postgres://user:p@ss!word@localhost:5432/db?sslmode=require"),
			},
			expected:    "Connection: postgres://user:p@ss!word@localhost:5432/db?sslmode=require",
			expectError: false,
		},
		{
			name:     "template with empty value",
			template: "Value: {{.Ref.empty}}",
			secretData: map[string][]byte{
				"empty": []byte(""),
			},
			expected:    "Value: ",
			expectError: false,
		},
		{
			name:     "template with multiline",
			template: "HOST={{.Ref.host}}\nPORT={{.Ref.port}}",
			secretData: map[string][]byte{
				"host": []byte("localhost"),
				"port": []byte("3306"),
			},
			expected:    "HOST=localhost\nPORT=3306",
			expectError: false,
		},
		{
			name:     "JSON template",
			template: `{"username":"{{.Ref.username}}","password":"{{.Ref.password}}"}`,
			secretData: map[string][]byte{
				"username": []byte("admin"),
				"password": []byte("secret"),
			},
			expected:    `{"username":"admin","password":"secret"}`,
			expectError: false,
		},
		{
			name:        "empty template string",
			template:    "",
			secretData:  map[string][]byte{"key": []byte("value")},
			expectError: true,
			errorMsg:    "template string cannot be empty",
		},
		{
			name:        "invalid template syntax",
			template:    "{{.Ref.username",
			secretData:  map[string][]byte{"username": []byte("john")},
			expectError: true,
			errorMsg:    "failed to parse template",
		},
		{
			name:        "missing key in template",
			template:    "Hello {{.Ref.missing}}",
			secretData:  map[string][]byte{"username": []byte("john")},
			expected:    "Hello <no value>",
			expectError: false,
		},
		{
			name:        "empty secret data",
			template:    "{{if .Ref.key}}Has key{{else}}No key{{end}}",
			secretData:  map[string][]byte{},
			expected:    "No key",
			expectError: false,
		},
		{
			name:        "nil secret data",
			template:    "Static text",
			secretData:  nil,
			expected:    "Static text",
			expectError: false,
		},
		{
			name:     "template with base64-like content",
			template: "Token: {{.Ref.token}}",
			secretData: map[string][]byte{
				"token": []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"),
			},
			expected:    "Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectError: false,
		},
		{
			name:     "template with printf functions",
			template: `{{printf "%s:%s@%s:%s/%s" .Ref.user .Ref.pass .Ref.host .Ref.port .Ref.db}}`,
			secretData: map[string][]byte{
				"user": []byte("dbuser"),
				"pass": []byte("dbpass"),
				"host": []byte("localhost"),
				"port": []byte("5432"),
				"db":   []byte("mydb"),
			},
			expected:    "dbuser:dbpass@localhost:5432/mydb",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderTemplate(tt.template, tt.secretData)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, string(result))
			}
		})
	}
}

func TestTemplateData(t *testing.T) {
	t.Run("template data structure", func(t *testing.T) {
		secretData := map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte("value2"),
		}

		template := "{{.Ref.key1}}-{{.Ref.key2}}"
		result, err := RenderTemplate(template, secretData)

		require.NoError(t, err)
		assert.Equal(t, "value1-value2", string(result))
	})
}

func TestRenderTemplate_ComplexScenarios(t *testing.T) {
	t.Run("environment file template", func(t *testing.T) {
		template := `DATABASE_HOST={{.Ref.db_host}}
DATABASE_PORT={{.Ref.db_port}}
DATABASE_USER={{.Ref.db_user}}
DATABASE_PASSWORD={{.Ref.db_password}}
DATABASE_NAME={{.Ref.db_name}}`

		secretData := map[string][]byte{
			"db_host":     []byte("postgres.example.com"),
			"db_port":     []byte("5432"),
			"db_user":     []byte("app_user"),
			"db_password": []byte("secure_pass_123"),
			"db_name":     []byte("application_db"),
		}

		result, err := RenderTemplate(template, secretData)
		require.NoError(t, err)

		expected := `DATABASE_HOST=postgres.example.com
DATABASE_PORT=5432
DATABASE_USER=app_user
DATABASE_PASSWORD=secure_pass_123
DATABASE_NAME=application_db`

		assert.Equal(t, expected, string(result))
	})

	t.Run("connection string template", func(t *testing.T) {
		template := "postgresql://{{.Ref.username}}:{{.Ref.password}}@{{.Ref.host}}:{{.Ref.port}}/{{.Ref.database}}?sslmode={{.Ref.sslmode}}"

		secretData := map[string][]byte{
			"username": []byte("myuser"),
			"password": []byte("mypassword"),
			"host":     []byte("db.example.com"),
			"port":     []byte("5432"),
			"database": []byte("mydb"),
			"sslmode":  []byte("require"),
		}

		result, err := RenderTemplate(template, secretData)
		require.NoError(t, err)

		expected := "postgresql://myuser:mypassword@db.example.com:5432/mydb?sslmode=require"
		assert.Equal(t, expected, string(result))
	})

	t.Run("yaml template", func(t *testing.T) {
		template := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  database_url: "{{.Ref.db_url}}"
  api_key: "{{.Ref.api_key}}"`

		secretData := map[string][]byte{
			"db_url":  []byte("postgres://localhost:5432/db"),
			"api_key": []byte("sk-1234567890abcdef"),
		}

		result, err := RenderTemplate(template, secretData)
		require.NoError(t, err)

		expected := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  database_url: "postgres://localhost:5432/db"
  api_key: "sk-1234567890abcdef"`

		assert.Equal(t, expected, string(result))
	})
}
