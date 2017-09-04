# Load Testing libraries
This is for running a perf .csv file agains a REST or S3 fileserver

## mkLoadTestFiles
Create files, locally on the target machine or via S3

## runLoadTest
Run the test 

## loadConfig
load the api key and secret from the same config file the
application we're testing uses. Peculiar to this app, but
the same as is used by the unit test and the app itself.

There are old ane new versions, identified by _older_ and Newer_
prefixes on their names. Newer starts 10 additional goroutines
in each period until we're past the target number.
Oddly, this seems to behave poorly: under investigation
