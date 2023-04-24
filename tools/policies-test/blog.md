---
share: true
created: Wednesday-19th-Apr-2023, 15:24
tags: blogs
---

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
