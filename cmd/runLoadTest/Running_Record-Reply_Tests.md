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
let's assume the old system was running at 200 requests a second, 
and was returning the average file in 0.3 second.

## Doing a smoke test
At this point, we can try a simple test. The classic debugging test
is
```
runLoadTest -v --rest  --tps 1 --for 1 \
	../load.csv http://calvin
```
That runs a single operation in verbose mode, which should look like
```
#yyy-mm-dd hh:mm:ss latency xfertime thinktime bytes url rc
2017/11/11 21:11:20 runLoadTest.go:194: starting, at 1 requests/second
2017/11/11 21:11:20 runLoadTest.go:137: Loaded 1 records, closing input
2017/11/11 21:11:22 restOps.go:189: 
Request: 
GET /zaphod-beebelbrox.jpg HTTP/1.1
Host: calvin
User-Agent: Go-http-client/1.1
Cache-Control: no-cache
Accept-Encoding: gzip

Response headers:
    Length: 122944
    Status code: 200 È OK
    Last-Modified : [Fri, 11 Aug 2017 13:59:57 GMT]
    Accept-Ranges : [bytes]
    Server : [nginx/1.10.3 (Ubuntu)]
    Content-Type : [image/jpeg]
    Content-Length : [12530]
    Date : [Sun, 12 Nov 2017 02:11:47 GMT]
    Connection : [keep-alive]
    Etag : ["598db85d-30f2"]
Response contents: 
HTTP/1.1 200 OK
Content-Length: 122944
Accept-Ranges: bytes
Connection: keep-alive
Content-Type: image/jpeg
Date: Sun, 12 Nov 2017 02:11:47 GMT
Etag: "598db85d-30f2"
Last-Modified: Fri, 11 Aug 2017 13:59:57 GMT
Server: nginx/1.10.3 (Ubuntu)

Body:
 ���'���OJ�����cDe��*�7;

```
followed by many lines of gibberish from viewing a gif as text.


## Debugging the system under test  
The next step is to replay from beginning to end without errors.

Instead of `--for 1`, we run through the whole file at some convenient
speed. If the system is expected to handle 100 request/second (TPS), 
try running at `--tps 100 --crash`, and see if you can get a clean run 
from beginning to end.

Any error will put the verbose switch on, and --crash will stop
as soon as there is an error, instead of continuing.

If you're not use the load tester is behaving properly, ass the `-d` 
debug option, and it will add extra information to the output.

You may have to take some problematic operations out of the input file, 
such as a get that always returns a 408 (a timeout), but be careful:
you might take something important out.


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
