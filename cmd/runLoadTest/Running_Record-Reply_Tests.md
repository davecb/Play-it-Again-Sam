# Running Record-Replay Tests

Replaying the exact log of requests and replies from a production 
system is an excellent way to do an accurate measurement of how 
a new system will behave.

The idea here is to increase the load in requests per second on
your system to see where it start to turn into a hockey-stick "_/" 
curve from too much load
 
![image](https://user-images.githubusercontent.com/559505/32694390-8e1bec20-c70c-11e7-9c5b-9da23b237b84.png)

In the image above, the latency time (delay, slowness) greatly increases
after 250 requests per second. Below that it gently increases, as we'll 
see later, from a quite pleasant 0.4s at 40 requests/second to an 
ugly 1.25 seconds at 120.

From that, we can conclude that the particular disk I'm testing is good 
only at low loads, but doesn't "hit the wall" until it's terribly overloaded.
Which is reasonable, as the disk is actually one designed for low 
power consumption and noise. That it can handle a huge "normal overload"
is a lovely thing to discover!

## Collecting initial data

We started this by finding a log from a server we wanted to replace, and
put it into a standard format.

The server was one of a group providing "object" storage via rest 
calls to an https server.  Because it was a web server, the access logs
record every request, and timestamp them. That instantly allows us to
calculate how many requests per second it was serving. 

The log, from nginx (or apache) looked like this
```
10.0.100.42 - - [09/Nov/2017:13:12:44 -0500] "GET /zaphod-beebelbrox.jpg 
HTTP/1.1" 200 122944 "-" "Dalvik/1.6.0 (Linux; U; Android 4.2.2; SM-T110 
Build/JDQ39)"
```
We converted it into a convenient format that we could read and write, 
making it easy to compare the original data with data form later tests.

This format looks like
```bash
#date       time     stats bytes  file                  rc  op
09/Nov/2017 13:12:44 0 0 0 122944 zaphod-beebelbrox.jpg 200 GET

```
The load tester will write a similar log, with a different date and time
and with the stats filled in.


## Adding response times
If you have access to the web server, you can add response times to
the log: they're calculated, but not reported for some reason.

That allows you to know not just how many request per second the 
existing system was seeing, but also how long it took to serve 
the average object at that level of load.

In the case of nginx, we would add "$request_time" as the last entry
in the log line, and copy it into the "latency" column of the stats
when putting it into our standard format.

For our  purposes,
let's assume the old system was running at 200 requests a asecond, 
and was returning the average file in 0.3 second.

## Doing a smoke test
At this point, we can try a simple test.   XXX
-for 1 -tps 1 -v

## Debugging the system under test  
-d to debug the load tester
-v to get more info about the SUT
all errors trigger -v
--crash to stop

## Doing a load test
try a large range
-tps 1000 -progress 100
plot 2
get an idea of the range you care about
plot 3

## Understanding what you're seeing
times
looking for the wait time to start to dominate
very approximately 2 RT
example of 0.1 to 0.3 as normal overload
