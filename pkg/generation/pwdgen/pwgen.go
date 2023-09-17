package pwdgen

import (
	"math"
	"math/rand"

	password "github.com/sethvargo/go-password/password"

	v1 "github.com/containerinfra/kube-secrets-operator/api/v1"
)

// GeneratePasswords generates a series of random string based on the supplied password templates
func GeneratePasswords(passwordSpec *v1.SecretTemplate) map[string][]byte {
	data := map[string][]byte{}

	for _, item := range passwordSpec.Data {
		// plain text value
		if item.Value != "" {
			data[item.Name] = []byte(item.Value)
			continue
		}
		passwordLength := getPasswordLength(&item)
		generatedPassword, err := password.Generate(passwordLength, getNumberOfDigits(&item), getNumberOfSymbols(&item), item.NoUpper, !item.NoRepeat)
		if err != nil {
			panic(err)
		}
		data[item.Name] = []byte(generatedPassword)
	}
	return data
}

func getPasswordLength(item *v1.SecretValueItemTemplate) int {
	lengthOfPassword := item.Length
	if item.MaxLength > 0 {
		lengthOfPassword = uint32(getRandomNumberBetween(int(item.MinLength), int(item.MaxLength)))
	} else {
		lengthOfPassword = uint32(math.Max(float64(lengthOfPassword), float64(item.MinLength)))
	}
	return int(lengthOfPassword)
}

func getNumberOfSymbols(item *v1.SecretValueItemTemplate) int {
	return int(math.Min(float64(getPasswordLength(item)), float64(getRandomNumberBetween(0, int(item.MaxSymbols)))))
}

func getNumberOfDigits(item *v1.SecretValueItemTemplate) int {
	return int(math.Min(float64(getPasswordLength(item)), float64(getRandomNumberBetween(0, int(item.MaxDigits)))))
}

func getRandomNumberBetween(min int, max int) int {
	if max == 0 {
		return 0
	} else if min >= max {
		return max
	}

	return min + int(rand.Int31n(int32(max-min)))
}
