# Results

## How to reproduces

Prerequisite:

- Vagrant

Make sure you have started both 0.13.0 and 0.14.0 instances or you can choose to start only one of them. Refer to [Deployment](../vagrant/README.md) for more details.

## Deploy 0.13.0

### E2E Test

While inside the 0.13.0 instance, run the following commands:

1. Go to the e2e test directory

   ```bash
   cd /home/vagrant/mainflux/tools/e2e
   ```

2. Build the e2e test binary

   ```bash
   make all
   ```

3. Run the e2e test

   ```bash
   hyperfine --prepare 'echo "delete from things; delete from channels; delete from connections;" | docker exec -i mainflux-things-db psql -U mainflux -d things && echo "delete from users;" | docker exec -i mainflux-users-db psql -U mainflux -d users && echo "delete from groups;" | docker exec -i mainflux-auth-db psql -U mainflux -d auth' './e2e --num 10 --num_of_messages 10' --export-json ../../../13e2e.json -r 10
   ```

   The above command will run the e2e test 10 times and export the result to `13e2e.json` file linked [here](13e2e.json).

### Fine-Grained Access Control

For fine-grained access control, version 0.13.0 does not support it. This is why we are shifting to version 0.14.0.

## Deploy 0.14.0

### E2E Test

While inside the 0.13.0 instance, run the following commands:

1. Go to the e2e test directory

   ```bash
   cd /home/vagrant/mainflux/tools/e2e
   ```

2. Build the e2e test binary

   ```bash
   make all
   ```

3. Run the e2e test

   ```bash
   hyperfine --prepare 'echo "delete from policies; delete from groups; delete from clients;" | docker exec -i mainflux-things-db psql -U mainflux -d things && echo "delete from policies; delete from groups; delete from clients;" | docker exec -i mainflux-users-db psql -U mainflux -d users' './e2e --num 10 --num_of_messages 10' --export-json ../../../14e2e.json -r 10
   ```

   The above command will run the e2e test 10 times and export the result to `13e2e.json` file linked [here](13e2e.json).

### Fine-Grained Access Control

While inside the 0.14.0 instance, run the following commands:

1. Go to the policies-test directory

   ```bash
   cd /home/vagrant/mainflux/tools/policies
   ```

2. Build the policies-test binary

   ```bash
   make all
   ```

3. Run the e2e test

   ```bash
   hyperfine --prepare 'echo "delete from policies; delete from groups; delete from clients;" | docker exec -i mainflux-users-db psql -U mainflux -d users' './policies-test' --export-json ../../../14.json
   ```

   The above command will run the e2e test 10 times and export the result to `13e2e.json` file linked [here](13e2e.json).

## Results

### E2E Test

#### 0.13.0

```json
{
    "mean": 5.66738176292,
    "stddev": 0.38708639298088476,
    "median": 5.63242686302,
    "user": 1.8688646600000003,
    "system": 2.09574456,
    "min": 5.11942044752,
    "max": 6.22411095652
}
```

#### 0.14.0

```json
{
    "mean": 10.691313304800001,
    "stddev": 0.8269843987467643,
    "median": 10.844986333200001,
    "user": 5.081039,
    "system": 6.12106696,
    "min": 8.9026274897,
    "max": 11.5588154477
}
```
