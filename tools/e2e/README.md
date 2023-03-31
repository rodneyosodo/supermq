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
created user with token eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODAyNzM5ODIsImlhdCI6MTY4MDI3MzA4MiwiaWRlbnRpdHkiOiJlMmUtcGF0aWVudC1mbG93ZXJAZW1haWwuY29tIiwiaXNzIjoiY2xpZW50cy5hdXRoIiwic3ViIjoiODI4MjJjNzUtY2I1NS00NzI4LTkyZTItMGE4MDA3YjE1MzNkIiwidHlwZSI6ImFjY2VzcyJ9.vk8sugrY2A8vUd4P9WPVCp0DgMH4jzyivc-ZSuY1RNWDo_DF5-XPN7hP4GR_pLYT_nvQBUwLj4nqHZcR-9rN_A
created users of ids:  [82af36e8-b4c1-40b3-a130-7dbc1442ef37 24e89346-f6df-4037-800d-1a6c9436d5e2 e4c604d9-9168-4305-a807-42fc44e06b57 6881d502-9210-4795-9c52-6e0276df124f 562fcc19-dd79-4c5f-b03e-899c628d5355 d602916f-a47a-4bb1-a541-b65a36438da0 a7727b63-c3d4-46b9-941a-e6cac0a6236c 490d1b53-2fce-4d70-bc93-6c0806cb0e1d e0826308-02c2-416f-a9db-67d9eb6d948b 69a15127-50d2-4a1c-aef2-929c8b7efea7]
created groups of ids:  [882caf92-e980-4f39-b9ad-56199ab216ef 2792cbbf-cdca-4f06-87b3-dce9f906e7a3 1d560201-6922-451e-868d-23ea31202f89 89735f4d-ca5a-4202-b32a-41977ccd44e6 73cbd542-492c-4958-9ecd-b2f758860342 c38d0498-d6df-4be5-ad09-aeccf31e64ac 795f6282-3db7-4eb4-83e5-ff356fafac9e 103f7bf2-b868-4aa6-963a-173595842738 3adb55a2-b0fc-4ae0-97fe-59994c5b6446 a17d32b7-0af1-4502-9ef2-95542b81ce04]
created things of ids:  [967c63e7-73a5-4a42-bca6-9fe0474a2028 2851c197-bd95-4e5f-b5b4-539ff6bf154a 286cd8c4-a224-4be6-b697-bff276096e0c 7feb937f-a7be-451c-b989-9fc95b3c46a3 f065b9a5-765d-42dd-8abd-979a87bb1715 d1d89e77-6b41-42ea-afa5-38bbed17b9e9 14449ffc-7930-4416-9c56-a568c6205445 167af38c-d521-4950-abea-02cf115a2a94 e5aaa368-1fe1-4d4d-bc03-031b6ad9bedc ee97d22e-6f70-486f-8ef0-af4c246bc083]
created channels of ids:  [900d0ce9-20c9-486f-8cab-adf6ad895561 b76d04ce-df65-4b98-a038-9d2a7892d9a9 d4c2bf77-5c38-4f63-a22a-6f556f86d463 c060c71b-a897-49ea-8342-20123efb77db 30c86c79-aec2-45f3-bd5d-2c831f9b71d9 51b8d31d-d873-4636-af0b-9a8875eed42f 5095d63b-7c72-430a-b06f-1e35ed4d3486 9254d627-2160-40db-bba4-9543c5f9547a 47603c14-26cd-45b8-9995-b96edd7dc8b5 8db3f336-b8d3-43b6-b741-d9f53b6040cb]
created policies for users, groups, things and channels
viewed users, groups, things and channels
updated users, groups, things and channels
sent and received messages from channels 
```
