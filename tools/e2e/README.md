# Mainflux Users Groups Things and Channels E2E Testing Tool

A simple utility to create a list of groups and users connected to these groups and channels and things connected to these channels.

## Installation

```bash
cd tools/e2e
make
```

### Usage

```bash
./e2e --help
Tool for testing end to end flow of mainflux by creating groups and users and assigning the together andcreating channels and things and connecting them together.
Complete documentation is available at https://docs.mainflux.io

Usage:
  e2e [flags]

Flags:
  -h, --help                   help for e2e
  -H, --host string            address for mainflux instance (default "localhost")
  -n, --num uint               number of users, groups, channels and things to create and connect (default 10)
  -N, --num_of_messages uint   number of messages to send (default 10)
  -p, --prefix string          name prefix for users, groups, things and channels
```

To use `-H` option, you can specify the address for the Mainflux instance as an argument when running the program. For example, if the Mainflux instance is running on another computer with the IP address 192.168.0.1, you could use the following command:

```bash
go run tools/e2e/cmd/main.go --host 142.93.118.47
```

This will tell the program to connect to the Mainflux instance running on the specified IP address.

If you want to create a list of channels with certificates:

```bash
go run tools/e2e/cmd/main.go --host localhost --num 10 --num_of_messages 100 --prefix e2e
```

Example of output:

```bash
created user with token eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODEyMDQ4ODIsImlhdCI6MTY4MTIwMzk4MiwiaWRlbnRpdHkiOiJlMmUtYmlsbG93aW5nLWdsaXR0ZXJAZW1haWwuY29tIiwiaXNzIjoiY2xpZW50cy5hdXRoIiwic3ViIjoiMTRlOTNmYmMtOGM2NC00Y2VmLWJjOTAtMzVhNzY4OGMxNGRmIiwidHlwZSI6ImFjY2VzcyJ9.MgmRGQyfVuALxazSdlvSpfknpUygLRBNikqPddm2_A0W6_ejhbmKx6_4YaekJc7q_MIOWwyu2dcbVMTof48Q2A
created users of ids:
         3bf53687-f332-4ba0-8d54-fc35051bf59a,
         ee737c0c-902b-4d23-b6b6-3e9294215078,
         8d0daca6-74f4-4358-a933-237b167820f4,
         2ec92797-3fab-4375-ad08-31c6abea42ff,
         a58d55ff-cc04-4748-bfdc-f612829670c8
created groups of ids:
         1f503500-e2a4-4c59-92d2-670312813e94,
         7839aed9-0696-4e99-98e1-0896ec630fb0,
         9889a07e-6e6f-42a0-b9ff-ccfd3f62b5e9,
         72bcf0b3-0361-463b-b1ed-cdd6556748d0,
         624e1670-0d47-40ff-b746-9442cdad82f3
created things of ids:
         69aefa74-96c5-4c67-b06c-c43f4e7eb3ba,
         29a71d9c-8574-44af-a8a3-8a42ae34eedd,
         aeeaa6d3-61b0-4e88-8118-0e2ddffab43f,
         f374967a-9488-440e-b88d-d5fb1e0ed684,
         c42550e2-6b50-470b-901d-9ca0cae956ee
created channels of ids:
         a72899ea-8394-4fa6-a534-5619e2f7e2d4,
         8f9dfb41-c11b-46e1-a340-ea1d4da20ebc,
         93e8e4f9-a7e1-4888-80a6-c50d424c0968,
         21c21fb9-f4e4-476f-947e-8858a192b003,
         96f7acee-fcc6-413f-ad70-aa95f5422e2a
created policies for users, groups, things and channels
viewed users, groups, things and channels
updated users, groups, things and channels
sent and received messages from channels
```
