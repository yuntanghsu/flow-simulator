# flow-simulator

## Getting started

### What's in this repository?

We have a flow-simulator that can generate normal flow and DDoS flow.

### Parameters

#### ddosTimeDuration int
    maximum time length for the ddos flow, unit is millisecond, default is 50
`./bin/main -ddosTimeDuration 10 -numNormalFlow 1 -numDDoSFlow 5`
![img.png](images%2Fimg.png)

`./bin/main -ddosTimeDuration 1000 -numNormalFlow 1 -numDDoSFlow 5`
![Screenshot 2024-01-05 at 4.00.03 PM.png](images%2FScreenshot%202024-01-05%20at%204.00.03%20PM.png)

#### numAttack int
    number of attack times, defaut is 1
`./bin/main -numNormalFlow 1 -numDDoSFlow 10 -numAttack 2 ` 

We will have two sets of attack scenario. In this case, we 
will have 10 attacks flow for destip=10.0.0.3 and 10.0.0.12

![Screenshot 2024-01-05 at 4.02.59 PM.png](images%2FScreenshot%202024-01-05%20at%204.02.59%20PM.png)

#### numAttackPods int
    number of DDoS attack Pods, default is 5
It affects the number of sourceIP will be used to attack a certain destinationIP. In above case, it is 5.

#### numDDoSFlow int
    number of ddos flow for each attach, default is 0

In previous example, we can see there are 10 flows for each attack.

#### numNormalFlow int
    number of normal flow, default is 100

#### numberOfPods int
    number of Pods in the cluster, default is 20

It means how many IPs can be used for sourceIP and destinationIP.

#### tcpState string
    TCP state for attack flow, default is TIME_WAIT

We can assign a specific TCP state to all DDoS flow.

`./bin/main -numNormalFlow 1 -numDDoSFlow 10 -numAttack 2 -tcpState "SYN"`

![Screenshot 2024-01-05 at 4.07.57 PM.png](images%2FScreenshot%202024-01-05%20at%204.07.57%20PM.png)

#### timeRange int
    time range for the start time, unit is minute, default is 10

`./bin/main -numNormalFlow 10 -timeRange 1`

![Screenshot 2024-01-05 at 4.12.32 PM.png](images%2FScreenshot%202024-01-05%20at%204.12.32%20PM.png)

`./bin/main -numNormalFlow 10 -timeRange 10`

![Screenshot 2024-01-05 at 4.13.32 PM.png](images%2FScreenshot%202024-01-05%20at%204.13.32%20PM.png)

### Running the programs

run `make bin` to build the binary file.
run `./bin/main` to start to generate flows.

Sample cmd:
`./bin/main -timeRange 30 -numNormalFlow 200 -numDDoSFlow 100 -numAttack 2`
