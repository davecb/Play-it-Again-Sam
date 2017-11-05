# mkLoadTestFiles(1) 
perf2seconds - report 1-second samples
## SYNOPSIS
Usage: perf2seconds file |-

## DESCRIPTION
This program reads a journal/file from runLoadsTest and creates a 
file with the values averages ove a one-second sample period.
The url and all subsequent field are replaces with a TPS field. 

## FILES
The input and output files are similar, of the form
```csv
#yyy-mm-dd hh:mm:ss latency xfertime thinktime bytes url rc op
2017-09-21 08:15:07.270 0 0 0 0 /upload/images/383bcc59-354b-46fb-b66c-0907b21fad94_albert.jpg 200 GET

```
The input fields are averaged to create an output record for each second,
with the URL files replaced by a TPS field.
 

## "SEE ALSO"
mkLoadTestFiles(1), nginx2perf(1), runLoadTest(1)

## EXAMPLES


## BUGS

## DIAGNOSTICS
none

## AUTHOR

David Collier-Brown
