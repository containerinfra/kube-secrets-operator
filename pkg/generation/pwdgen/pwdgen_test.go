package pwdgen

import (
	"testing"

	v1 "github.com/containerinfra/kube-secrets-operator/api/v1"
)

func TestGeneratePasswords(t *testing.T) {
	data := GeneratePasswords(&v1.SecretTemplate{
		Data: []v1.SecretValueItemTemplate{
			{
				Name:   "MY_PASSWORD",
				Length: 10,
			},
		},
	})
	if data == nil {
		t.Errorf("generatePasswords did not return any output")
	}

	if len(data) != 1 {
		t.Errorf("Amount of items in data from generatePasswords was incorrect, got: %d, want: %d", len(data), 1)
	}

	password := string(data["MY_PASSWORD"])
	if len(password) != 10 {
		t.Errorf("Length of generated password was incorrect, got: %d, want: %d", len(password), 10)
	}
}

func TestGeneratePasswordsDifferent(t *testing.T) {
	data := GeneratePasswords(&v1.SecretTemplate{
		Data: []v1.SecretValueItemTemplate{
			{
				Name:   "MY_PASSWORD",
				Length: 10,
			},
		},
	})

	data2 := GeneratePasswords(&v1.SecretTemplate{
		Data: []v1.SecretValueItemTemplate{
			{
				Name:   "MY_PASSWORD",
				Length: 10,
			},
		},
	})
	password := string(data["MY_PASSWORD"])
	password2 := string(data2["MY_PASSWORD"])

	if password == password2 {
		t.Errorf("Password should be regenered, got: %s and : %s", password, password2)
	}

}

func TestGeneratePasswordsRandomLength(t *testing.T) {

	data := GeneratePasswords(&v1.SecretTemplate{
		Data: []v1.SecretValueItemTemplate{
			{
				Name:      "MY_PASSWORD",
				MinLength: 10,
				MaxLength: 32,
			},
		},
	})
	password := string(data["MY_PASSWORD"])

	if len(password) < 10 {
		t.Errorf("Length of generated password was incorrect, got: %d, wanted a min of: %d", len(password), 10)
	}

	if len(password) > 32 {
		t.Errorf("Length of generated password was incorrect, got: %d, wanted a max of: %d", len(password), 32)
	}
}

func TestGeneratePasswordsRandomLengthDifferent(t *testing.T) {

	for count := 2; count < 100; count++ {
		data := GeneratePasswords(&v1.SecretTemplate{
			Data: []v1.SecretValueItemTemplate{
				{
					Name:      "MY_PASSWORD",
					MinLength: 10,
					MaxLength: uint32(count),
				},
			},
		})
		password := string(data["MY_PASSWORD"])

		if count >= 10 {
			if len(password) < 10 {
				t.Errorf("Length of generated password was incorrect, got: %d, wanted a min of: %d", len(password), 10)
			}
		}

		if len(password) > count {
			t.Errorf("Length of generated password was incorrect, got: %d, wanted a max of: %d", len(password), count)
		}
	}

}

func TestGeneratePasswordsMaxLengthHasPreference(t *testing.T) {
	data := GeneratePasswords(&v1.SecretTemplate{
		Data: []v1.SecretValueItemTemplate{
			{
				Name:      "MY_PASSWORD",
				MinLength: 10,
				MaxLength: 2,
			},
		},
	})
	password := string(data["MY_PASSWORD"])
	if len(password) != 2 {
		t.Errorf("Length of generated password was incorrect, got: %d, wanted: %d", len(password), 2)
	}
}
