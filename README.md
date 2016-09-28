# mcafee-bruteforce
Brute force password cracker for McAfee VirusScan

Brute Force Password Cracker for McAfee VirusScan
=======================
This is a Go program that attempts to recover the password used by McAfee VirusScan to unlock the UI by brute forcing it. 
--------------------------------------------------------------------------------------------
**Created by Valer Bocan, PhD (www.bocan.ro)**

**How to brute force McAfee VirusScan Console password:**
- **STEP 1**: Retrieve the hash of the McAfee VirusScan hash from *HKEY_LOCAL_MACHINE\SOFTWARE\WOW6432Node\McAfee\DesktopProtection\UPIEx*
The hash should look like a BASE64 encoded string, such as 1Or2ZtCTFvnWGxR1M1OnPV+88Eg=
- **STEP 2**: Set up cracking parameters: hash to crack (from previous step), password length, alphabet and degree of parallelism
- **STEP 3**: Launch program on the fastest machine you have available
- **STEP 4**: ... wait ... zzz ...

NOTE: On a laptop with an Intel i7 processor and a parallelism degree of 40, the speed is approximately 1.4M ops/sec.

Some hashes for you to crack:

- 1Or2ZtCTFvnWGxR1M1OnPV+88Eg= (4 chars, all lowercase, embarrasingly simple)
- rGZD24f1qaTAD1H4YusGRAh7YSQ= (5 chars, lower and upper case characters)
- s0YloBD4DBXzfd5bVSl/Tu6uLAs= (5 chars, lower and upper case characters plus one digit)
- +EQwmmJeZonna+XpXmzj4AGCg14= (6 chars, lower and upper case characters)
- aGS9k7AF8J2/Tmaa4yn0acZayCs= (6 chars, lower and upper case characters plus one digit)
- aGS9k7AF8J2/Tmaa4yn0acZayCs= (6 chars, lower case characters, one digit and mathematical symbols)
- w3YDvNwMWmcNqmntfzse8oFOHEM= (7 chars, digits, one upper case and one lower case letter plus a dash)
- WoEuoXP/I1hxbgri8otQ+ZSCh6A= (9 chars, lower and upper case letters)
- +EQwmmJeZonna+XpXmzj4AGCg14= (extremely long password, presumably 17 characters long. This is a real challenge!)

**NOTE:** This program is of little practical value because of the time it may take to find long and complicated passwords. A more sensible approach is to use precompiled word lists, leet speak rules and other variations to reduce the time.
