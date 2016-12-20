# Example tasks

[This](tasks/task-file-disk.yml) example task will collect metrics from the disk collector.

## Running the example

### Requirements 
 * `docker` and `docker-compose` are **installed** and **configured** 

Running the sample is as *easy* as running the script `./run-file-disk.sh`. 

Note: If you want to run the example without going through Docker you could 
run `file-disk.sh`.  

## Files

- [run-file-disk.sh](run-file-disk.sh) 
    - The example is launched with this script
- [tasks/task-file-disk.yml](tasks/task-file-disk.yml)
    - Snap task definition
- [docker-compose.yml](docker-compose.yml)
    - A docker compose file which defines a snapteld container
        - "runner" is the container where snapteld is run from.  You will be dumped 
        into a shell in this container after running 
        [run-file-disk.sh](run-file-disk.sh).  Exiting the shell will
        trigger cleaning up the containers used in the example.
- [file-disk.sh](file-disk.sh)
    - Downloads `snapteld`, `snaptel`, `snap-plugin-collector-disk`, and starts the task [tasks/task-file-disk.yml](tasks/task-file-disk.yml).
