package core

import "sync"

var (
	validators     = []string{"Validator1", "Validator2", "Validator3"}
	currentIndex   = 0
	validatorMutex = sync.Mutex{}
)

func GetNextValidator() string {
	validatorMutex.Lock()
	defer validatorMutex.Unlock()

	currentValidator := validators[currentIndex]

	currentIndex = (1 + currentIndex) % len(validators)

	return currentValidator
}
