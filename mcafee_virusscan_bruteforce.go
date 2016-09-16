package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// How to brute force McAfee VirusScan Console password
// STEP 1: Retrieve the hash of the McAfee VirusScan hash from HKEY_LOCAL_MACHINE\SOFTWARE\WOW6432Node\McAfee\DesktopProtection\UPIEx
// The hash should look a BASE64 encoded string, such as 1Or2ZtCTFvnWGxR1M1OnPV+88Eg=
// STEP 2: Set up cracking parameters in the section below: hash to crack (from previous step), password length, alphabet and degree of parallelism
// STEP 3: Launch program on the fastest machine you have available
// STEP 4: ... wait ... zzz ...
//
// NOTE: On a laptop with an Intel i7 processor and a parallelism degree of 40, the speed is approximately 1.4M ops/sec.

// Some hashes for you to crack:
//    1Or2ZtCTFvnWGxR1M1OnPV+88Eg= (4 chars, all lowercase, embarrasingly simple)
//	  rGZD24f1qaTAD1H4YusGRAh7YSQ= (5 chars, lower and upper case characters)
//    s0YloBD4DBXzfd5bVSl/Tu6uLAs= (5 chars, lower and upper case characters plus one digit)
//    +EQwmmJeZonna+XpXmzj4AGCg14= (6 chars, lower and upper case characters)
//    aGS9k7AF8J2/Tmaa4yn0acZayCs= (6 chars, lower and upper case characters plus one digit)
//    aGS9k7AF8J2/Tmaa4yn0acZayCs= (6 chars, lower case characters, one digit and mathematical symbols)
//    w3YDvNwMWmcNqmntfzse8oFOHEM= (7 chars, digits, one upper case and one lower case letter plus a dash)
//    WoEuoXP/I1hxbgri8otQ+ZSCh6A= (9 chars, lower and upper case letters)
//    +EQwmmJeZonna+XpXmzj4AGCg14= (extremely long password, presumably 17 characters long. This is a real challenge!)

const (
	hashToCrack         = "1Or2ZtCTFvnWGxR1M1OnPV+88Eg=" // Choose your target
	minPasswordLength   = 1                              // Minimum password length
	maxPasswordLength   = 6                              // Maximum password length
	parallelismDegree   = 50                             // Number of goroutines to launch for hashing
	alphabet            = alphabetMixAlphaNumeric        // Alphabet selection
	enableHTTPReporting = false                          // Put here true if you want to monitor progress via HTTP (on port 8888, just use a browser)
)

// Common alphabets
const (
	alphabetNumeric            = `0123456789`
	alphabetAlphaNumeric       = `ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`
	alphabetLowerAlpha         = `abcdefghijklmnopqrstuvwxyz`
	alphabetLowerAlphaNumeric  = `abcdefghijklmnopqrstuvwxyz0123456789`
	alphabetMixAlpha           = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`
	alphabetMixAlphaNumeric    = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`
	alphabetStandardUSKeyboard = ` !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_` + "`" + `abcdefghijklmnopqrstuvwxyz{|}~`
)

var (
	currentSpeed         string
	currentPassword      string
	currentPasswordMutex = &sync.Mutex{}
	timeAtStart          time.Time
	operationCount       uint64
)

func main() {
	// Print welcome screen
	fmt.Println("McAfee VirusScan Brute Force Password Cracker v1.0 - September 2016 (Valer BOCAN, PhD <valer@bocan.ro>)")
	fmt.Printf("Attempting attack on hash '%s'.\n", hashToCrack)
	fmt.Printf("Using alphabet '%s' (%d chars).\n", alphabet, len(alphabet))
	fmt.Printf("Passwords range: %d to %d chars.\n", minPasswordLength, maxPasswordLength)
	fmt.Println()

	// Launch statistics reporting via HTTP
	if enableHTTPReporting {
		// Set up listening on port 8888
		go func() {
			http.HandleFunc("/", printStatistics)    // set router
			err := http.ListenAndServe(":8888", nil) // set listen port
			if err != nil {
				log.Fatal("Error setting up statistics: ", err)
			}
		}()
	}

	// Record the start time
	timeAtStart = time.Now()

	// Launch goroutine that displays the current speed
	go func() {
		for {
			// Determine current speed
			atomic.StoreUint64(&operationCount, 0)
			time.Sleep(1 * time.Second)
			ops := atomic.LoadUint64(&operationCount)
			// Scale value to something meaningful for the user
			switch {
			case ops > 1000000000:
				currentSpeed = fmt.Sprintf("Speed: %.2fB ops/sec.", float64(ops)/1000000000.0)
			case ops > 1000000:
				currentSpeed = fmt.Sprintf("Speed: %.2fM ops/sec.", float64(ops)/1000000.0)
			case ops > 1000:
				currentSpeed = fmt.Sprintf("Speed: %.0fK ops/sec.", float64(ops)/1000.0)
			default:
				currentSpeed = fmt.Sprintf("Speed: %d ops/sec.", ops)
			}
			// Get the time that has elapsed
			elapsed := time.Since(timeAtStart)
			fmt.Printf("[%v] %v Password: [%s]\n", elapsed, currentSpeed, currentPassword)
		}
	}()

	// Generate and process passwords
	passwords := generatePasswords(alphabet, minPasswordLength, maxPasswordLength)
	foundPassword := processPasswords(passwords)

	// Record the stop time
	elapsed := time.Since(timeAtStart)

	// Display running results
	if foundPassword != "" {
		fmt.Printf("Password is [%s] and was found in %s.\n", foundPassword, elapsed)
	} else {
		fmt.Printf("Password was not found after running %s.\n", elapsed)
	}

	fmt.Println("Done.")
}

// printStatistics prints current cracking progress to console
func printStatistics(w http.ResponseWriter, r *http.Request) {
	elapsed := time.Since(timeAtStart)
	fmt.Fprintf(w, "[%v] %v Password: [%s]\n", elapsed, currentSpeed, currentPassword)
}

// processPasswords reads passwords from the input string channel, applies the hash function
// and compares with the target hash. This function spawns a specified number of goroutines
// and blocks until processing is done.
func processPasswords(passwords <-chan string) string {
	finishChan := make(chan string, parallelismDegree)
	// Start multiple goroutines to consume the passwords generated via the input channel
	for i := 1; i <= parallelismDegree; i++ {
		go func(id int) {
			for passwd := range passwords {
				//fmt.Printf("Goroutine #%d: %v\n", id, passwd)
				hash := computeMcAfeeHash(passwd)
				// Check whether the computed hash matches what we want to test against
				if hash == hashToCrack {
					finishChan <- passwd // Finished, password found
					return
				}
			}
			finishChan <- "" // Finished, password not found
		}(i)
	}

	// Wait for all goroutines to finish
	var finalPassword string
	for i := 0; i < parallelismDegree; i++ {
		v := <-finishChan
		if v != "" {
			// Password was found by a goroutine, skip the wait and exit the program
			finalPassword = v
			break
		}
	}
	// The finalPassword variable contains the password found during the search or a null string if nothing was found
	return finalPassword
}

// ---------------------------------------------------------------------------------------------------
// Password generation (recursive algorithm)
func generatePasswords(alphabet string, minLength int, maxLength int) <-chan string {
	chn := make(chan string)
	go func(chn chan string) {
		defer close(chn)
		for l := minLength; l <= maxLength; l++ {
			addLetter(chn, "", alphabet, l)
		}
	}(chn)
	return chn
}

func addLetter(chn chan string, combo string, alphabet string, length int) {
	if length == 0 {
		// We have reached the desired password length, send it over the channel
		chn <- combo
		// Increase the operation count (so we can determine the speed in another goroutine)
		atomic.AddUint64(&operationCount, 1)
		// Also store the current password so we can display on demand
		currentPasswordMutex.Lock()
		currentPassword = combo
		currentPasswordMutex.Unlock()
		return
	}
	// Add another character to the combo
	for _, ch := range alphabet {
		newCombo := combo + string(ch)
		addLetter(chn, newCombo, alphabet, length-1)
	}
}

// ---------------------------------------------------------------------------------------------------
// Hash computation for McAfee VirusScan Console
// Retrieve the hash from Windows registry at: HKEY_LOCAL_MACHINE\SOFTWARE\WOW6432Node\McAfee\DesktopProtection\UPIEx
// Hashing schema is: Base64(SHA1(UTF-16("\x01\x0f\x0d\x33" + password)))
// Exmple: password "test" would generate 1Or2ZtCTFvnWGxR1M1OnPV+88Eg=
// Challenge: try to crack hash +EQwmmJeZonna+XpXmzj4AGCg14=

const mcAfeePadding = "\x01\x0f\x0d\x33"

func computeMcAfeeHash(password string) string {
	buffer := make([]byte, 0, 32)
	paddedPassword := mcAfeePadding + password
	for _, c := range paddedPassword {
		lsb := byte(c & 0x00ff)
		msb := byte(c & 0xff00 >> 8)
		buffer = append(buffer, lsb, msb)
	}
	// Compute SHA1
	bufferSHA1 := sha1.Sum(buffer)
	// If we wanted to encode result in Base64, we would do like this
	hash := base64.StdEncoding.EncodeToString(bufferSHA1[:])
	return hash
}
