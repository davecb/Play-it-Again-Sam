# mkLoadTestFiles(1) 
mkLoadTestFiles - create files to get in a test
## SYNOPSIS
Usage: mkLoadTestFiles [-from N -for N][-v] load-file.csv url

## DESCRIPTION
This program creates a set of files for a load test, by default in a 
local filesystem.  The files contain different amounts of the same
sequence of random data.

It exists to avoid having the logic in runLoadTest, although it reports
its performance in exactly the same way as runLoadTest does.

### Data options   
-for int 
* number of records to use, eg 1000.   
  This limits the number of files to be created.  

-from int 
* number of records to skip, eg 100.   
  This starts at a particular record in the file
  
-zero
* create zero-size files, to avoid measuring transfer time during the test.


### Misc options      
     
-v
* add verbose messages    
  This is for debugging the system where the files are being created.

### Config-file options 
These options are from the config-file parser, which allows any of the
above options to be specified in a configuration file.
   
-allowMissingConfig 
 * Don't terminate the app if the ini file cannot be read. 
   
-allowUnknownFlags 
 * Don't terminate the app if ini file contains unknown flags.  
 
-config string 
 * Path to ini config for using in go flags. May be relative to the 
 current executable path.   
 
-configUpdateInterval duration 
* Update interval for re-reading config file set via -config flag. 
  Zero disables config file re-reading. 
   
-dumpflags 
* Dumps values for all flags defined in the app into stdout in 
  ini-compatible syntax and terminates the app.    


## FILES
The input files are the same as those used by runLoadTest, of the form
```csv
#yyy-mm-dd hh:mm:ss latency xfertime thinktime bytes url rc op
2017-09-21 08:15:07.270 0 0 0 0 /zaphod-beeblebrox.jpg 200 GET
```
As an input, only the url and the file size are significant. The url is 
concatenated to the prefix provide on the command-line and used as the 
filename to be created.
 

## "SEE ALSO"
perf2seconds.md, nginx2perf.md, runLoadTest.md, Running_Record-Reply_Tests.md

## EXAMPLES


## BUGS

## DIAGNOSTICS
If an error occurs, the program reports as much as possible and stops.

During normal operation, a small number of status messages will also
be written to stderr to indicate progress.  


## AUTHOR

David Collier-Brown
