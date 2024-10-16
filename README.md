# Distributed Data Store using Kademlia 
This is a lab done in the course D7024E Mobile and Distributed Computing systems at Luleå University of Tech-
nology by:
- Elvira Forslund Widenroth - elvfor-0@student.ltu.se
- Jenn Sundström - ejyuso-2@student.ltu.se
  
Read the report, D7024E_Kademlia_Group7.pdf, for an in depth description of the system.  
## Scripts
#### ./up.sh 
Creates Image and containers and rebuilds new code on them

#### ./stop.sh
Stops containers and cleans up unused images

#### ./down.sh
Removes containers and cleans up unused images
## Prerequisites
- Install docker: https://docs.docker.com/engine/install/
- Install GO: https://go.dev/doc/install

## Running the code
1. Clone or download the code.
2. Start docker. 
3. Change directory to the "scripts" folder.
4. run ./up.sh in terminal.
5. Enter a specific node using docker attach 'container id or name', example docker attach d7024e-kadlab-kademliaNodes-1. 
6. When inside a node, use PUT, GET and EXIT as described in the rapport. 

## Testing the code

To run all test with test coverage run: go test --cover ./... 

