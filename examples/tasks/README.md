# Example tasks

[This](task-mock-disk.yml) example task will collect metrics from the disk collector.

## Running the example

### Requirements 
 * `docker` and `docker-compose` are **installed** and **configured** 

Running the sample is as *easy* as running the script `./run-mock-disk.sh`. 

Note: If you want to run the example without going through Docker you could 
run `mock-disk.sh`.  

## Files

- [run-mock-disk.sh](run-mock-disk.sh) 
    - The example is launched with this script
- [task-mock-disk.yml](task-mock-disk.yml)
    - Snap task definition
- [docker-compose.yml](docker-compose.yml)
    - A docker compose file which defines a snapteld container
        - "runner" is the container where snapteld is run from.  You will be dumped 
        into a shell in this container after running 
        [run-mock-disk.sh](run-mock-disk.sh).  Exiting the shell will
        trigger cleaning up the containers used in the example.
- [mock-disk.sh](mock-disk.sh)
    - Downloads `snapteld`, `snaptel`, `snap-plugin-collector-disk`, and starts the task [task-mock-disk.yml](task-mock-disk.yml).
