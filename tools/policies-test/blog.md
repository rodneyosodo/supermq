# ❓Showcase Clients

In today's digital age, the term "client" can refer to a wide range of entities, including human users, devices, and software applications. Managing these clients can be a complex and challenging task, especially in large-scale systems. In this document, we will explore the concept of clients in detail and discuss how they are managed in a microservices-based architecture. Specifically, we will take a deep dive into the Clients service, a central component of our platform, and examine how it as the Mainflux Things and Users services to provide a single, cohesive solution for managing clients. We will also explore the Clients service's core features, including client authentication, access control policies, and traceability, to gain a comprehensive understanding of how it fits into the broader system architecture. By the end of this post, you will have a solid grasp of the Clients service's functionality and how it enables efficient and secure client management.

## 🎯 Objective

1. Client Authorization
2. Client Authentication
3. Traceability

## 1️⃣ Client Authorization

In today's digital world, businesses need to have a robust system in place to manage access control and ensure security. Access control is crucial in preventing unauthorized access to sensitive data and resources, and it has become even more important with the increasing number of cyber attacks.

In the past, we had been using the Keto Rules Engine to manage our access control policies. While it had served us well, we found that it lacked some essential features that we needed to ensure maximum security. After careful consideration, we decided to switch to a new approach - Fine-Grained Access Control.

In this document, we will explore the reasons why we made this switch, the implications of this change, and the benefits we have experienced so far. We will also discuss some of the tests and metrics we used to evaluate the effectiveness of the new system and why we believe it is superior to the old one.

By the end, you should have a clear understanding of why we decided to switch to Fine-Grained Access Control and how it works. So let's dive in and explore the world of access control together.

### Keto

[Keto rules engine][keto-github-url] is an open-source software designed to enforce access control policies in cloud-native environments. It was developed by the [ORY community][ory-github-url] to provide authentication, authorization, access control, application network security, and delegation for cloud-native applications.

With Ory Permissions, you have access to a contemporary authorization system that can accommodate any application, regardless of the size or complexity of the required access-control lists (ACLs). This system is based on the Ory Keto Permission Server, an open-source software, and is the first implementation of the design principles and specifications outlined in [Zanzibar, which is Google's consistent, global authorization system][zanzibar-paper].

Ory Keto utilizes a system of regulations and principles to decide if a particular demand should be approved or declined. It maintains records of connections between subjects and objects, like a user's link to a file or a user's inclusion in a group, and organizes them into namespaces to create a diagram of entity relationships.

Upon receiving a request, Ory Keto evaluates the requested operation alongside the associated subject and object. It navigates the graph created from the stored data to recognize the connections between the entities and applies appropriate rules and policies to determine whether to allow or deny the request.

### Zanzibar

Zanzibar is a distributed authorization system that manages access control policies for hundreds of client applications across more than 1,500 namespaces. It uses a relational data model to store access control policies as a set of relation tuples that define which principals (e.g., users or services) are allowed to perform which actions on which resources.

Zanzibar provides a simple, flexible, and expressive policy language that allows administrators to define fine-grained access control policies based on attributes of the principal, resource, and context. Policies can be composed using boolean logic and can include time-based constraints and other conditions.

When a client application needs to perform an authorization check, it sends a request to Zanzibar with the principal, action, and resource as input. Zanzibar evaluates the request against the relevant access control policy and returns a response indicating whether the action is allowed or denied.

To handle the high query load from multiple client applications, Zanzibar distributes the load across multiple servers organized in several dozen clusters around the world. It also uses caching and de-duplication mechanisms to reduce latency and improve performance.

Zanzibar provides two types of requests: Safe and Recent. Safe requests have zookies more than 10 seconds old and can be served within the region most of the time, while Recent requests have zookies less than 10 seconds old and often require interregion round trips. Zanzibar uses this distinction to minimize latency and improve availability.

### Limitations of Keto

While the keto rules engine has its benefits, there are some drawbacks that can limit its effectiveness in certain situations. These include:

1. Difficulty: Setting up and configuring the keto rules engine can be challenging, which may deter some organizations from using it. It also requires technical expertise to maintain and manage.
2. Scalability: If the number of rules and policies increases, the performance of the keto rules engine may decrease. This can lead to slower response times and a less efficient system.
3. Lack of dynamic policy support: The keto rules engine is built for static policies and may not be suitable for allowing access rights to be granted or revoked based on context.
4. Need for greater control: Traditional access control methods based on roles and permissions may not provide enough control over data access. Fine-grained access control can be beneficial by providing more specific and detailed control over who can access what data.
5. Improved security: Fine-grained access control can increase the security of a system by allowing administrators to set specific access permissions for different users or groups. This can help prevent unauthorized access and minimize the risk of data breaches.

### Implications

- [x] Increased security: Fine-grained access control provides a higher level of security by allowing administrators to limit access to specific resources and actions, reducing the risk of unauthorized access.
- [x] More complex administration: Implementing fine-grained access control can be more complex than using coarse-grained access control. It requires more granular permission management and additional configuration steps.
- [x] Increased performance overhead: Fine-grained access control may result in increased performance overhead due to the additional checks and controls that need to be performed for each access request.

## 2️⃣ Client Authentication

Many applications and services are going to the cloud in today's world, and protecting our online identity is more critical than ever. As a result, access tokens and refresh tokens have grown in popularity as a means of authenticating users in cloud-based systems.

Access tokens function similarly to keys in that they give access to certain resources or services. When you log in to a service, you are given an access token that allows you to utilize the resources that you are permitted to use. The token provides information about your identity as well as your degree of access.

Refresh tokens are used to extend the life of your access token. Access tokens typically expire after a specific period of time, however a refresh token can produce a new access token, allowing you to continue using the service without having to log in again.

Using access and refresh tokens allows you to access sites without continually checking in or inputting your credentials. It also guarantees that only authorized users have access to the resources and that their activities are tracked.

Example:

```bash
curl 'http://localhost/users/tokens/issue' -H 'Content-Type: application/json' -H 'Accept: application/json' --data-raw '{"identity": "example@email.com", "secret": "12345678"}'

HTTP/1.1 201 Created
Access-Control-Allow-Headers: *
Access-Control-Allow-Methods: *
Access-Control-Allow-Origin: *
Connection: keep-alive
Content-Length: 707
Content-Type: application/json
Date: Mon, 24 Apr 2023 13:50:56 GMT
Server: nginx/1.23.3
Strict-Transport-Security: max-age=63072000; includeSubdomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY

{
    "access_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODIzOTgyNTYsImlhdCI6MTY4MjM0NDI1NiwiaWRlbnRpdHkiOiJleGFtcGxlQGVtYWlsLmNvbSIsImlzcyI6ImNsaWVudHMuYXV0aCIsInN1YiI6ImYzYWU3MGJiLWU1NzctNDMyYi1iOGI4LTMyZGNkNjYxYTc4YSIsInR5cGUiOiJhY2Nlc3MifQ.Ow5RW-0_08FSsuehUrytPuVWgF0FiVQ3lYVaZC2lpDtzqFWZ0GFfiF4nw6v8lzDKA15L2pGV7C3BvIY2jvnq1A",
    "access_type": "Bearer",
    "refresh_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODI0MzA2NTYsImlhdCI6MTY4MjM0NDI1NiwiaWRlbnRpdHkiOiJleGFtcGxlQGVtYWlsLmNvbSIsImlzcyI6ImNsaWVudHMuYXV0aCIsInN1YiI6ImYzYWU3MGJiLWU1NzctNDMyYi1iOGI4LTMyZGNkNjYxYTc4YSIsInR5cGUiOiJyZWZyZXNoIn0.fiq035MFgrce_fWoIlt9KPc0hlb214T8ypIQfMRy-S1k4LWXFUxxKwOAdzOjiyDT3VhkwA8Vxq8Ts7Uo1Ex1ZA"
}
```

## 3️⃣ Traceability

![Tracing][tracing-screenshot]

## 4️⃣ Implementation

The access privileges are executed in the following format `S-A-O` - Subject-Action-Object Where Subject is Client ID, Action is one of the supported actions (read, write, update, delete, list or add), and Object is **Client ID** or **Group ID**.

```go
// Policy represents an argument struct for making a policy related function calls.

type Policy struct {
 OwnerID   string    `json:"owner_id"`
 Subject   string    `json:"subject"`
 Object    string    `json:"object"`
 Actions   []string  `json:"actions"`
 CreatedAt time.Time `json:"created_at"`
 UpdatedAt time.Time `json:"updated_at"`
}

var examplePolicy = Policy{
 Subject: userToken,
 Object:  groupID,
 Actions: []string{groupListAction},
}
```

Policies handling initial implementation is meant to be used on the **Group** level.

There are three types of policies:

- M Policy for messages
- G Policy for Group rights
- C Policy for Clients that are group members

M Policy represents client rights to send and receive messages to a group. Only group members with corresponding rights can publish or receive messages to/from the group.

G Policy represents client rights to modify the group itself. Only group members with correct rights can modify or update the group, or add/remove members to/from the group.

Finally, the C policy represents rights the member has over other members of the group.

The rules are specified in the **policies** association table. The table looks like this:

| subject | object | actions                                     |
| ------- | ------ | ------------------------------------------- |
| clientA | groupA | ["g_add", "g_list", "g_update", "g_delete"] |
| clientB | groupA | ["c_list", "c_update", "c_delete"]          |
| clientC | groupA | ["c_update"]                                |
| clientD | groupA | ["c_list"]                                  |
| clientE | groupB | ["c_list", "c_update", "c_delete"]          |
| clientF | groupB | ["c_update"]                                |
| clientD | groupB | ["c_list"]                                  |
| clientG | groupC | ["m_read"]                                  |
| clientH | groupC | ["m_read", "m_write"]                       |

Actions such as `c_list`, and `c_update` represent actions allowed for the client with client_id to execute over all the other clients that are members of the group with gorup_id. Actions such as `g_update` represent actions allowed for the client with client_id to execute against a group with group_id.

For the sake of simplicity, all the operations at the moment are executed on the **group level** - the group acts as a namespace in the context of authorization and is required.

1. Actions for `clientA`

   - when `clientA` list groups `groupA` will be listed
   - they are able to add members to `groupA`
   - they are able to update `groupA`
   - they are able to change the status of `groupA`

2. Actions for `clientB`

   - when the list clients they will list `clientA`, `clientC` and `clientD` since they are connected in the same group `groupA` and they have `c_list` actions
   - they are able to update clients connected to the same group they are connect in i.e they are able to update `clientA`, `clientC` and `clientD` since they are in the same `groupA`
   - they are able to change clients status of clients connected to the same group they are connect in i.e they are able to change the status of `clientA`, `clientC` and `clientD` since they are in the same group `groupA`

3. Actions for `clientC`

   - they are able to update clients connected to the same group they are connect in i.e they are able to update `clientA`, `clientB` and `clientD` since they are in the same `groupA`

4. Actions for `clientD`

   - when the list clients they will list `clientA`, `clientB` and `clientC` since they are connected in the same group `groupA` and they have `c_list` actions and also `clientE` and `clientF` since the are connected to the same group `groupB` and they have `c_list` actions

5. Actions for `clientE`

   - when the list clients they will list `clientF` and `clientD` since they are connected in the same group `groupB` and they have `c_list` actions
   - they are able to update clients connected to the same group they are connect in i.e they are able to update `clientF` and `clientD` since they are in the same `groupB`

6. Actions for `clientF`

   - they are able to update clients connected to the same group they are connect in i.e they are able to update `clientE`, and `clientD` since they are in the same `groupB`

7. Actions for `clientG`

   - they are able to read messages posted in group `groupC`

8. Actions for `clientH`

   - they are able to read from `groupC` and write messages to `groupC`

## 5️⃣ Demo

### Getting Started

We will be using [this clients branch][client-branch-url]

1. Clone Repository

   ```bash
   git clone https://github.com/mainflux/mainflux
   cd mainflux
   ```

2. Checkout to branch

   ```bash
   git remote add rodneyosodo https://github.com/rodneyosodo/mainflux.git
   git fetch rodneyosodo BLOG-showcaseclients0140
   git switch --track rodneyosodo/BLOG-showcaseclients0140
   ```

3. Start mainflux

   ```bash
   make all && make dockers
   make run
   ```

4. Run Policies Test Tool

    ```bash
    go run tools/policies-test/cmd/main.go
    ```

    This creates the users and policies demonstrated by this table:

    | subject | object | actions                                     |
    | ------- | ------ | ------------------------------------------- |
    | clientA | groupA | ["g_add", "g_list", "g_update", "g_delete"] |
    | clientB | groupA | ["c_list", "c_update", "c_delete"]          |
    | clientC | groupA | ["c_update"]                                |
    | clientD | groupA | ["c_list"]                                  |
    | clientE | groupB | ["c_list", "c_update", "c_delete"]          |
    | clientF | groupB | ["c_update"]                                |
    | clientD | groupB | ["c_list"]                                  |
    | clientG | groupC | ["m_read"]                                  |
    | clientH | groupC | ["m_read", "m_write"]                       |

    The output should look like this:

    ```bash
    created user with token eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODIzOTUxNzcsImlhdCI6MTY4MjM0MTE3NywiaWRlbnRpdHkiOiJjbGllbnRhZG1pbkBwb2xpY2llcy5jb20iLCJpc3MiOiJjbGllbnRzLmF1dGgiLCJzdWIiOiJjNDhlMTc0Ni02MjVlLTRmODctODNiYS04Y2MyYjE0YmNjOTEiLCJ0eXBlIjoiYWNjZXNzIn0.vFLebBWAustsGVBoCuoBjMltdKNXxoL7cNIFGPEzpqXfs0DZXxA1fEoTf4p2NYjxn0qwfYvfaf_MHFb6hhrtOQ
    created users of ids:
    2bdb18f7-d513-4378-b0b8-3fbc6e7c03ec
    8a994f84-9355-43e4-8017-d43c7bca964b
    58035137-f4d7-4492-88cf-ea36650215b9
    0b9e08a7-cd01-4e4a-91bf-d7edea2926df
    21581a96-54f9-4126-9811-e87b5550d444
    1ff4adf0-3b4a-4a2a-9882-02bbf9e89d15
    created groups of ids:
    5ce09d3b-9930-49b1-ada7-6248798b652b
    b4d366af-8d08-4168-a9da-e0b8cdefb162
    created policies for users, groups
    validated policies for client a
    validated policies for client b
    validated policies for client c
    validated policies for client d
    validated policies for client e
    validated policies for client f
    ```

5. Test the Api using the [Postman Collection][showcase-clients-postman-collection-url]

benefits we have experienced so far
tests and metrics we used to evaluate the effectiveness of the new system

### Notable Features

This section describes the most important features of the Mainflux system introduced in this release. Other features are described in the [mainflux docs][mainflux-docs-url].

#### 1️⃣ Users Service

##### Create User

To start working with the Mainflux system, you need to create a user account.

> Must-have: e-mail (this must be unique as it identifies the user) and password (password must contain at least 8 characters) and status
> Optional: name, tags, metadata, status and role

```bash
curl -s -S -i -X POST -H "Content-Type: application/json" http://localhost/users -d '{"credentials": {"identity": "<user_email>", "secret": "<user_password>"}}'
```

For example:

```bash
curl -s -S -i -X POST -H "Content-Type: application/json" http://localhost/users -d '{"credentials": {"identity": "john.doe@email.com", "secret": "12345678"}, "name": "John Doe", "status": "enabled"}'

HTTP/1.1 201 Created
Server: nginx/1.23.3
Date: Wed, 05 Apr 2023 09:36:36 GMT
Content-Type: application/json
Content-Length: 227
Connection: keep-alive
Location: /users/f7c55a1f-dde8-4880-9796-b3a0cd05745b
Strict-Transport-Security: max-age=63072000; includeSubdomains
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: *
Access-Control-Allow-Headers: *

{
    "id": "f7c55a1f-dde8-4880-9796-b3a0cd05745b",
    "name": "John Doe",
    "credentials": { "identity": "john.doe@email.com", "secret": "" },
    "created_at": "2023-04-05T09:36:36.40605Z",
    "updated_at": "2023-04-05T09:36:36.40605Z",
    "status": "enabled"
}
```

You can also use <user_token> so that the new user has an owner for example:

```bash
curl -s -S -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $USER_TOKEN" http://localhost/users -d '{"credentials": {"identity": "john.doe2@email.com", "secret": "12345678"}}'

HTTP/1.1 201 Created
Server: nginx/1.23.3
Date: Wed, 05 Apr 2023 09:51:30 GMT
Content-Type: application/json
Content-Length: 259
Connection: keep-alive
Location: /users/838dc3bb-fac9-4b39-ba4c-97a051c7c10d
Strict-Transport-Security: max-age=63072000; includeSubdomains
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: *
Access-Control-Allow-Headers: *

{
    "id": "838dc3bb-fac9-4b39-ba4c-97a051c7c10d",
    "owner": "f7c55a1f-dde8-4880-9796-b3a0cd05745b",
    "credentials": { "identity": "john.doe2@email.com", "secret": "" },
    "created_at": "2023-04-05T09:51:30.404336Z",
    "updated_at": "2023-04-05T09:51:30.404336Z",
    "status": "enabled"
}
```

##### Create Token

To log in to the Mainflux system, you need to create a `user_token`.

> Must-have: registered e-mail and password

```bash
curl -s -S -i -X POST -H "Content-Type: application/json" http://localhost/users/tokens/issue -d '{"identity":"<user_email>",
"secret":"<user_password>"}'
```

For example:

```bash
curl -s -S -i -X POST -H "Content-Type: application/json" http://localhost/users/tokens/issue -d '{"identity": "john.doe@email.com", "secret": "12345678"}'

HTTP/1.1 201 Created
Server: nginx/1.23.3
Date: Wed, 05 Apr 2023 09:43:04 GMT
Content-Type: application/json
Content-Length: 709
Connection: keep-alive
Strict-Transport-Security: max-age=63072000; includeSubdomains
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: *
Access-Control-Allow-Headers: *

{
    "access_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODA2ODg2ODQsImlhdCI6MTY4MDY4Nzc4NCwiaWRlbnRpdHkiOiJqb2huLmRvZUBlbWFpbC5jb20iLCJpc3MiOiJjbGllbnRzLmF1dGgiLCJzdWIiOiJmN2M1NWExZi1kZGU4LTQ4ODAtOTc5Ni1iM2EwY2QwNTc0NWIiLCJ0eXBlIjoiYWNjZXNzIn0.IkRNbWm609z3FffXdDn8C0FdfsEtId0woKjRsVu0ZxFXSAWZhczbFpGzWIzj43Pw0NnPaR0qEl-Kckx6vXRreg",
    "refresh_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODA3NzQxODQsImlhdCI6MTY4MDY4Nzc4NCwiaWRlbnRpdHkiOiJqb2huLmRvZUBlbWFpbC5jb20iLCJpc3MiOiJjbGllbnRzLmF1dGgiLCJzdWIiOiJmN2M1NWExZi1kZGU4LTQ4ODAtOTc5Ni1iM2EwY2QwNTc0NWIiLCJ0eXBlIjoicmVmcmVzaCJ9.xRpa1F76PWkHsiD8pzAtzrwDNyKKGkL1j6Vwczhr0uKgbAd1pOwYNUQhh6XVYtzFM9S5MYxRpQYfYOkcubW1UQ",
    "access_type": "Bearer"
}
```

##### Refresh Token

To issue another access_token after getting expired, you need to use a `refresh_token`.

> Must-have: refresh_token

```bash
curl -s -S -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer <refresh_token>" http://localhost/users/tokens/refresh
```

For example:

```bash
curl -s -S -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $REFRESH_TOKEN" http://localhost/users/tokens/refresh

HTTP/1.1 201 Created
Server: nginx/1.23.3
Date: Wed, 05 Apr 2023 10:44:15 GMT
Content-Type: application/json
Content-Length: 709
Connection: keep-alive
Strict-Transport-Security: max-age=63072000; includeSubdomains
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: *
Access-Control-Allow-Headers: *

{
    "access_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODA2OTIzNTUsImlhdCI6MTY4MDY5MTQ1NSwiaWRlbnRpdHkiOiJqb2huLmRvZUBlbWFpbC5jb20iLCJpc3MiOiJjbGllbnRzLmF1dGgiLCJzdWIiOiJmN2M1NWExZi1kZGU4LTQ4ODAtOTc5Ni1iM2EwY2QwNTc0NWIiLCJ0eXBlIjoiYWNjZXNzIn0.1-YxZd-UhSAyvAN0OkGVZJLl7e2ahDr3AnsSWe_Yhz3w1Ua3Cl1p62NOelZAwm-0qgpZgT-GwI6KG-J7jq_UGQ",
    "refresh_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODA3Nzc4NTUsImlhdCI6MTY4MDY5MTQ1NSwiaWRlbnRpdHkiOiJqb2huLmRvZUBlbWFpbC5jb20iLCJpc3MiOiJjbGllbnRzLmF1dGgiLCJzdWIiOiJmN2M1NWExZi1kZGU4LTQ4ODAtOTc5Ni1iM2EwY2QwNTc0NWIiLCJ0eXBlIjoicmVmcmVzaCJ9.GcTF_g31GYpMDqUHOXbyOiG3WnfMCssKmLvM2ZjsSajBHdcdZ_JKn85vOUCKfF-thcFaWL5VXCTSHN11bKdObw",
    "access_type": "Bearer"
}
```

##### Get Users

You can get all users in the database by querying this endpoint. List all users request accepts limit, offset, email and metadata query parameters.

> Must-have: `user_token`

```bash
curl -s -S -i -X GET -H "Authorization: Bearer <user_token>" http://localhost/users
```

For example:

```bash
curl -s -S -i -X GET -H "Authorization: Bearer $USER_TOKEN" http://localhost/users

HTTP/1.1 200 OK
Server: nginx/1.23.3
Date: Wed, 05 Apr 2023 09:53:19 GMT
Content-Type: application/json
Content-Length: 285
Connection: keep-alive
Strict-Transport-Security: max-age=63072000; includeSubdomains
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: *
Access-Control-Allow-Headers: *

{
    "limit": 10,
    "total": 1,
    "users": [{
        "id": "838dc3bb-fac9-4b39-ba4c-97a051c7c10d",
        "owner": "f7c55a1f-dde8-4880-9796-b3a0cd05745b",
        "credentials": { "identity": "john.doe2@email.com", "secret": "" },
        "created_at": "2023-04-05T09:51:30.404336Z",
        "updated_at": "0001-01-01T00:00:00Z",
        "status": "enabled"
    }]
}
```

If you want to paginate your results then use this

> Must have: `user_token`
> Additional parameters: `offset`, `limit`, `metadata`, `name`, `identity`, `tag`, `status` and `visbility`

```bash
curl -s -S -i -X GET -H "Authorization: Bearer <user_token>" http://localhost/users?offset=<offset>&limit=<limit>&identity=<identity>
```

For example to get users you are in the same group with:

```bash
curl -s -S -i -X GET -H "Authorization: Bearer $USER_TOKEN" http://localhost/users?offset=0&limit=5&visibility=shared

HTTP/1.1 200 OK
Server: nginx/1.23.3
Date: Wed, 05 Apr 2023 09:57:11 GMT
Content-Type: application/json
Content-Length: 284
Connection: keep-alive
Strict-Transport-Security: max-age=63072000; includeSubdomains
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: *
Access-Control-Allow-Headers: *

{
    "limit": 10,
    "total": 3,
    "users": [
        {
            "id": "1ef7d142-2261-4b90-b967-17ed19bda551",
            "name": "updatedByC",
            "owner": "0ea52631-c7fc-4c17-ad17-103c48c127b7",
            "credentials": {
                "identity": "clienta@policies.com",
                "secret": ""
            },
            "created_at": "2023-04-25T10:08:34.504248Z",
            "updated_at": "0001-01-01T00:00:00Z",
            "status": "enabled"
        },
        {
            "id": "cdda125b-58d0-49ca-ac6b-92c122382565",
            "name": "updatedA",
            "owner": "0ea52631-c7fc-4c17-ad17-103c48c127b7",
            "credentials": {
                "identity": "clientc@policies.com",
                "secret": ""
            },
            "created_at": "2023-04-25T10:08:34.614491Z",
            "updated_at": "0001-01-01T00:00:00Z",
            "status": "enabled"
        },
        {
            "id": "d06fb7bb-efd4-4653-ab4b-f5b2762689d9",
            "name": "updatedByF",
            "owner": "0ea52631-c7fc-4c17-ad17-103c48c127b7",
            "credentials": {
                "identity": "clientd@policies.com",
                "secret": ""
            },
            "created_at": "2023-04-25T10:08:34.667833Z",
            "updated_at": "0001-01-01T00:00:00Z",
            "status": "enabled"
        }
    ]
}
```


##### Get Group Children

Get all group children, list requests accepts limit and offset query parameters

> Must-have: `user_token`

```bash
curl -s -S -i -X GET -H "Content-Type: application/json" -H "Authorization: Bearer <user_token>" http://localhost/groups/<group_id>/children
```

For example:

```bash
curl -s -S -i -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $USER_TOKEN" http://localhost/groups/3d5e2b97-5400-444e-a48e-eb5b35f8de75/children?tree=true

HTTP/1.1 200 OK
Server: nginx/1.23.3
Date: Wed, 05 Apr 2023 11:40:34 GMT
Content-Type: application/json
Content-Length: 782
Connection: keep-alive
Access-Control-Expose-Headers: Location

{
    "limit": 10,
    "total": 10,
    "groups": [
        {
            "id": "3d5e2b97-5400-444e-a48e-eb5b35f8de75",
            "owner_id": "1890d8b7-5020-4166-bdf0-53c15b2bfa8b",
            "name": "throbbing-bush",
            "metadata": {
                "Update": "icy-snow"
            },
            "path": "3d5e2b97-5400-444e-a48e-eb5b35f8de75",
            "children": [
                {
                    "id": "9acb42eb-09f6-46a7-ba46-19a469d2f417",
                    "owner_id": "1890d8b7-5020-4166-bdf0-53c15b2bfa8b",
                    "parent_id": "3d5e2b97-5400-444e-a48e-eb5b35f8de75",
                    "name": "crimson-river",
                    "metadata": {
                        "Update": "autumn-dawn"
                    },
                    "level": 1,
                    "path": "3d5e2b97-5400-444e-a48e-eb5b35f8de75.9acb42eb-09f6-46a7-ba46-19a469d2f417",
                    "children": [
                        {
                            "id": "01bd641f-4765-4ce7-a725-55bb96619f6d",
                            "owner_id": "1890d8b7-5020-4166-bdf0-53c15b2bfa8b",
                            "parent_id": "9acb42eb-09f6-46a7-ba46-19a469d2f417",
                            "name": "late-forest",
                            "metadata": {
                                "Update": "wandering-dust"
                            },
                            "level": 2,
                            "path": "3d5e2b97-5400-444e-a48e-eb5b35f8de75.9acb42eb-09f6-46a7-ba46-19a469d2f417.01bd641f-4765-4ce7-a725-55bb96619f6d",
                            "children": [
                                {
                                    "id": "1e7680f1-523d-49b5-870b-410444ea0193",
                                    "owner_id": "1890d8b7-5020-4166-bdf0-53c15b2bfa8b",
                                    "parent_id": "01bd641f-4765-4ce7-a725-55bb96619f6d",
                                    "name": "snowy-meadow",
                                    "metadata": {
                                        "Update": "dawn-fog"
                                    },
                                    "level": 3,
                                    "path": "3d5e2b97-5400-444e-a48e-eb5b35f8de75.9acb42eb-09f6-46a7-ba46-19a469d2f417.01bd641f-4765-4ce7-a725-55bb96619f6d.1e7680f1-523d-49b5-870b-410444ea0193",
                                    "children": [
                                        {
                                            "id": "2e87d4c8-d463-4931-852f-cd97d6b1b1a8",
                                            "owner_id": "1890d8b7-5020-4166-bdf0-53c15b2bfa8b",
                                            "parent_id": "1e7680f1-523d-49b5-870b-410444ea0193",
                                            "name": "restless-shape",
                                            "metadata": {
                                                "Update": "young-flower"
                                            },
                                            "level": 4,
                                            "path": "3d5e2b97-5400-444e-a48e-eb5b35f8de75.9acb42eb-09f6-46a7-ba46-19a469d2f417.01bd641f-4765-4ce7-a725-55bb96619f6d.1e7680f1-523d-49b5-870b-410444ea0193.2e87d4c8-d463-4931-852f-cd97d6b1b1a8",
                                            "children": [
                                                {
                                                    "id": "713c4561-c048-4424-a43f-ab319d78c5d4",
                                                    "owner_id": "1890d8b7-5020-4166-bdf0-53c15b2bfa8b",
                                                    "parent_id": "2e87d4c8-d463-4931-852f-cd97d6b1b1a8",
                                                    "name": "young-snow",
                                                    "metadata": {
                                                        "Update": "cool-field"
                                                    },
                                                    "level": 5,
                                                    "path": "3d5e2b97-5400-444e-a48e-eb5b35f8de75.9acb42eb-09f6-46a7-ba46-19a469d2f417.01bd641f-4765-4ce7-a725-55bb96619f6d.1e7680f1-523d-49b5-870b-410444ea0193.2e87d4c8-d463-4931-852f-cd97d6b1b1a8.713c4561-c048-4424-a43f-ab319d78c5d4",
                                                    "children": [
                                                        {
                                                            "id": "a82769cf-f395-4e28-84b9-d1a3e1cf2d08",
                                                            "owner_id": "1890d8b7-5020-4166-bdf0-53c15b2bfa8b",
                                                            "parent_id": "713c4561-c048-4424-a43f-ab319d78c5d4",
                                                            "name": "holy-cloud",
                                                            "metadata": {
                                                                "Update": "rough-brook"
                                                            },
                                                            "level": 6,
                                                            "path": "3d5e2b97-5400-444e-a48e-eb5b35f8de75.9acb42eb-09f6-46a7-ba46-19a469d2f417.01bd641f-4765-4ce7-a725-55bb96619f6d.1e7680f1-523d-49b5-870b-410444ea0193.2e87d4c8-d463-4931-852f-cd97d6b1b1a8.713c4561-c048-4424-a43f-ab319d78c5d4.a82769cf-f395-4e28-84b9-d1a3e1cf2d08",
                                                            "children": [
                                                                {
                                                                    "id": "041a9c6f-9d3c-4003-b1a5-33c1b7c2fb92",
                                                                    "owner_id": "1890d8b7-5020-4166-bdf0-53c15b2bfa8b",
                                                                    "parent_id": "a82769cf-f395-4e28-84b9-d1a3e1cf2d08",
                                                                    "name": "shy-sky",
                                                                    "metadata": {
                                                                        "Update": "frosty-waterfall"
                                                                    },
                                                                    "level": 7,
                                                                    "path": "3d5e2b97-5400-444e-a48e-eb5b35f8de75.9acb42eb-09f6-46a7-ba46-19a469d2f417.01bd641f-4765-4ce7-a725-55bb96619f6d.1e7680f1-523d-49b5-870b-410444ea0193.2e87d4c8-d463-4931-852f-cd97d6b1b1a8.713c4561-c048-4424-a43f-ab319d78c5d4.a82769cf-f395-4e28-84b9-d1a3e1cf2d08.041a9c6f-9d3c-4003-b1a5-33c1b7c2fb92",
                                                                    "children": [
                                                                        {
                                                                            "id": "79efdc79-6c45-41ee-9a5b-2be7f1966b8e",
                                                                            "owner_id": "1890d8b7-5020-4166-bdf0-53c15b2bfa8b",
                                                                            "parent_id": "041a9c6f-9d3c-4003-b1a5-33c1b7c2fb92",
                                                                            "name": "cold-dust",
                                                                            "metadata": {
                                                                                "Update": "late-tree"
                                                                            },
                                                                            "level": 8,
                                                                            "path": "3d5e2b97-5400-444e-a48e-eb5b35f8de75.9acb42eb-09f6-46a7-ba46-19a469d2f417.01bd641f-4765-4ce7-a725-55bb96619f6d.1e7680f1-523d-49b5-870b-410444ea0193.2e87d4c8-d463-4931-852f-cd97d6b1b1a8.713c4561-c048-4424-a43f-ab319d78c5d4.a82769cf-f395-4e28-84b9-d1a3e1cf2d08.041a9c6f-9d3c-4003-b1a5-33c1b7c2fb92.79efdc79-6c45-41ee-9a5b-2be7f1966b8e",
                                                                            "children": [
                                                                                {
                                                                                    "id": "7862b875-b882-42e8-8f07-60ff2bf36ae3",
                                                                                    "owner_id": "1890d8b7-5020-4166-bdf0-53c15b2bfa8b",
                                                                                    "parent_id": "79efdc79-6c45-41ee-9a5b-2be7f1966b8e",
                                                                                    "name": "crimson-violet",
                                                                                    "metadata": {
                                                                                        "Update": "nameless-wood"
                                                                                    },
                                                                                    "level": 9,
                                                                                    "path": "3d5e2b97-5400-444e-a48e-eb5b35f8de75.9acb42eb-09f6-46a7-ba46-19a469d2f417.01bd641f-4765-4ce7-a725-55bb96619f6d.1e7680f1-523d-49b5-870b-410444ea0193.2e87d4c8-d463-4931-852f-cd97d6b1b1a8.713c4561-c048-4424-a43f-ab319d78c5d4.a82769cf-f395-4e28-84b9-d1a3e1cf2d08.041a9c6f-9d3c-4003-b1a5-33c1b7c2fb92.79efdc79-6c45-41ee-9a5b-2be7f1966b8e.7862b875-b882-42e8-8f07-60ff2bf36ae3",
                                                                                    "created_at": "2023-04-25T10:20:06.590844Z",
                                                                                    "updated_at": "2023-04-25T10:20:07.252504Z",
                                                                                    "status": "enabled"
                                                                                }
                                                                            ],
                                                                            "created_at": "2023-04-25T10:20:06.586532Z",
                                                                            "updated_at": "2023-04-25T10:20:07.239465Z",
                                                                            "status": "enabled"
                                                                        }
                                                                    ],
                                                                    "created_at": "2023-04-25T10:20:06.582677Z",
                                                                    "updated_at": "2023-04-25T10:20:07.226618Z",
                                                                    "status": "enabled"
                                                                }
                                                            ],
                                                            "created_at": "2023-04-25T10:20:06.578263Z",
                                                            "updated_at": "2023-04-25T10:20:07.21341Z",
                                                            "status": "enabled"
                                                        }
                                                    ],
                                                    "created_at": "2023-04-25T10:20:06.573867Z",
                                                    "updated_at": "2023-04-25T10:20:07.200635Z",
                                                    "status": "enabled"
                                                }
                                            ],
                                            "created_at": "2023-04-25T10:20:06.570059Z",
                                            "updated_at": "2023-04-25T10:20:07.185478Z",
                                            "status": "enabled"
                                        }
                                    ],
                                    "created_at": "2023-04-25T10:20:06.566292Z",
                                    "updated_at": "2023-04-25T10:20:07.170515Z",
                                    "status": "enabled"
                                }
                            ],
                            "created_at": "2023-04-25T10:20:06.562134Z",
                            "updated_at": "2023-04-25T10:20:07.155872Z",
                            "status": "enabled"
                        }
                    ],
                    "created_at": "2023-04-25T10:20:06.557407Z",
                    "updated_at": "2023-04-25T10:20:07.141191Z",
                    "status": "enabled"
                }
            ],
            "created_at": "2023-04-25T10:20:06.552598Z",
            "updated_at": "2023-04-25T10:20:07.12637Z",
            "status": "enabled"
        }
    ]
}
```

#### 2️⃣ Things Service

##### Connect

Connect things to channels

> Must-have: `user_token`, `channel_id` and `thing_id`. Group IDs are channel IDs and client IDs are thing IDs.

```bash
curl -s -S -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer <user_token>" http://localhost/connect -d '{"group_ids": ["<channel_id>"], "client_ids": ["<thing_id>"]}'
```

For example:

```bash
curl -s -S -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $USER_TOKEN" http://localhost/connect -d '{"group_ids": ["830cf52a-4928-4cfb-af44-e82a696c9ac9"], "client_ids": ["ee6de1d4-aa1d-443a-9848-7857a90d03bd"]}'

HTTP/1.1 201 Created
Server: nginx/1.23.3
Date: Wed, 05 Apr 2023 15:35:29 GMT
Content-Type: application/json
Content-Length: 231
Connection: keep-alive
Access-Control-Expose-Headers: Location

{
    "policies": [{
        "owner_id": "",
        "subject": "ee6de1d4-aa1d-443a-9848-7857a90d03bd",
        "object": "830cf52a-4928-4cfb-af44-e82a696c9ac9",
        "actions": ["m_write", "m_read"],
        "created_at": "0001-01-01T00:00:00Z",
        "updated_at": "0001-01-01T00:00:00Z"
    }]
}
```

You can specify the action you want to allow for the thing to the channel by adding `actions` to the request body. Which will be `m_write` and `m_read` by default.

```bash
curl -s -S -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $USER_TOKEN" http://localhost/connect -d '{"group_ids": ["830cf52a-4928-4cfb-af44-e82a696c9ac9"], "client_ids": ["ee6de1d4-aa1d-443a-9848-7857a90d03bd"], actions: ["m_write"]}'

HTTP/1.1 201 Created
Server: nginx/1.23.3
Date: Wed, 05 Apr 2023 15:35:29 GMT
Content-Type: application/json
Content-Length: 231
Connection: keep-alive
Access-Control-Expose-Headers: Location

{
    "policies": [{
        "owner_id": "",
        "subject": "ee6de1d4-aa1d-443a-9848-7857a90d03bd",
        "object": "830cf52a-4928-4cfb-af44-e82a696c9ac9",
        "actions": ["m_write"],
        "created_at": "0001-01-01T00:00:00Z",
        "updated_at": "0001-01-01T00:00:00Z"
    }]
}
```

Connect thing to channel

> Must-have: `user_token`, `channel_id` and `thing_id`

```bash
curl -s -S -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer <user_token>" http://localhost/channels/<channel_id>/things/<thing_id>
```

For example:

```bash
curl -s -S -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $USER_TOKEN" http://localhost/channels/830cf52a-4928-4cfb-af44-e82a696c9ac9/things/977bbd33-5b59-4b7a-a9c3-8adda5d98255

HTTP/1.1 201 Created
Server: nginx/1.23.3
Date: Wed, 05 Apr 2023 15:37:01 GMT
Content-Type: application/json
Content-Length: 281
Connection: keep-alive
Access-Control-Expose-Headers: Location

{
    "policies": [{
        "owner_id": "f7c55a1f-dde8-4880-9796-b3a0cd05745b",
        "subject": "977bbd33-5b59-4b7a-a9c3-8adda5d98255",
        "object": "830cf52a-4928-4cfb-af44-e82a696c9ac9",
        "actions": ["m_write", "m_read"],
        "created_at": "2023-04-05T15:37:01.120301Z",
        "updated_at": "2023-04-05T15:37:01.120301Z"
    }]
}
```

##### Disconnect

Disconnect things from channels specified by lists of IDs.

> Must-have: `user_token`, `group_ids` and `client_ids`

```bash
curl -s -S -i -X PUT -H "Content-Type: application/json" -H "Authorization: Bearer <user_token>" http://localhost/disconnect -d '{"client_ids": ["<thing_id_1>", "<thing_id_2>"], "group_ids": ["<channel_id_1>", "<channel_id_2>"]}'
```

For example:

```bash
curl -s -S -i -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $USER_TOKEN" http://localhost/disconnect -d '{"group_ids": ["830cf52a-4928-4cfb-af44-e82a696c9ac9"], "client_ids": ["ee6de1d4-aa1d-443a-9848-7857a90d03bd"]}'

HTTP/1.1 204 No Content
Server: nginx/1.23.3
Date: Wed, 05 Apr 2023 15:39:19 GMT
Content-Type: application/json
Connection: keep-alive
Access-Control-Expose-Headers: Location
```

Disconnect thing from the channel

> Must-have: `user_token`, `channel_id` and `thing_id`

```bash
curl -s -S -i -X DELETE -H "Content-Type: application/json" -H "Authorization: Bearer <user_token>" http://localhost/channels/<channel_id>/things/<thing_id>
```

For example:

```bash
curl -s -S -i -X DELETE -H "Content-Type: application/json" -H "Authorization: Bearer $USER_TOKEN" http://localhost/channels/830cf52a-4928-4cfb-af44-e82a696c9ac9/things/977bbd33-5b59-4b7a-a9c3-8adda5d98255

HTTP/1.1 204 No Content
Server: nginx/1.23.3
Date: Wed, 05 Apr 2023 15:40:58 GMT
Content-Type: application/json
Connection: keep-alive
Access-Control-Expose-Headers: Location
```

## 6️⃣ Conclusion

Finally, fine-grained access control has various advantages that can improve a system's security and efficiency. It allows for more granular and thorough control over who may access what data, which can aid in preventing illegal access and reducing the risk of data breaches. Furthermore, it enables access privileges to be given or revoked based on context, increasing flexibility and agility.

## 7️⃣ References

1. [https://github.com/rodneyosodo/mainflux/tree/BLOG-0130to0140/_material](https://github.com/rodneyosodo/mainflux/tree/BLOG-0130to0140/_material)
2. [https://github.com/ory/keto](https://github.com/ory/keto)
3. [https://research.google/pubs/pub48190/](https://research.google/pubs/pub48190/)

[keto-github-url]: https://github.com/ory/keto
[ory-github-url]: https://github.com/ory
[zanzibar-paper]: https://research.google/pubs/pub48190/
[tracing-screenshot]: https://res.cloudinary.com/rodneyosodo/image/upload/v1682346170/blogs/showcase-clients/Screenshot_from_2023-04-24_16-59-39_i25uqy.png
[client-branch-url]: https://github.com/rodneyosodo/mainflux/tree/BLOG-showcaseclients0140
[showcase-clients-postman-collection-url]: https://ox6flab.postman.co/workspace/UltravioletClients~acba89f8-0255-435e-9547-59e542474e21/collection/5838754-17252174-4c7a-4fd8-b3a8-97033ca7267d?action=share&creator=5838754
[mainflux-docs-url]: https://docs.mainflux.io/api/
