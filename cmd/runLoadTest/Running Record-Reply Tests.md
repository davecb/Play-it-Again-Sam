# Running Record-Replay Tests


aim is to get "_/" curve and know where it stats to get bad
plot 0

## Collecting initial data
find a server log
convert it into a "round trip" format


## Adding response times
optional: shows you how well the existing system is doing
plot 1

## Doing a smoke test
-for 1 -tps 1 -v

## Debugging the system under test  \
-d to debug the load tester
-v to get more info about the SUT
all errors trigger -v
--save to see the file that is returned

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
