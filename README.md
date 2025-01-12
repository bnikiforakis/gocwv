# Go Microservice to write CRuX API response into a DB  
  
  
## Scope:  
The point of this repo, is to write the Main Core Web Vitals (as tracked for the past 20 days from google) to a table, called `crux_metrics`.  
It will run for all URLS found under `/urls.json`, and will keep only the `desiredMetrics`, atm set to be `largest_contentful_paint`, `cumulative_layout_shift`, `interaction_to_next_paint`  


## Requirements:
1. Go Installation
2. Path for go in the profile (mac)
3. Install sqlc `brew install sqlc`
4. install the lib pq `go get github.com/lib/pq`
5. Install Go migration package `brew install golang-migrate`


## Instructions:
1. Run docker (either app or from terminal)
2. Get docker image for container `docker pull postgres:17.2-alpine3.21`
3. On the root directory, you can run everything from Makefile:  
a. Run `make postgres` to spin a new container for postgres  
b. Run `make createdb` to create a db schema  
c. Run `make migrateup` to create the crux_metric table   
d. Run `make run` to run the service, and post the results to the DB
4. When done, you can run `make dropdb` to drop the DB, `make stopdocker` to stop the docker, and `make killdocker` to delete the container  
5. You can always check your docker image by `docker ps` or `docker ps -a` 


## Secrets: 
You can either use a .env file, if you want to run it locally, or you need to set up the Github Secrets in your project.   
In any case, the following are needed:
```CRUX_API_KEY="your crux api key"    
DB_HOST="Host of the database"    
DB_PORT="Port of the database"    
DB_USER="User of the database"    
DB_PASSWORD="Password for the user"    
DB_NAME="Name of the database"