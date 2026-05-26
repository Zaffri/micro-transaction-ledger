# Micro Transaction Ledger

An event-driven distributed banking ledger utilising microservices written in Go. User's can transfer money between accounts where operations are safely managed across backend services using a choreographed saga pattern (asynchronous messaging) ensuring data is rolledback in the event of a failure. Duplicate payments and processing is prevented using idempotency keys.


TODO: add screenshot/GIF of UI demo

## Basic requirements
- Payments between accounts
- Immutable ledger (audit trail)
- Simple transaction statuses with reasoning; Pending, Rejected_Fraud, Authorised
- Idempotency; prevention of payment duplication or processes
- Demo UI for sending payments and for end-to end visualisation purposes
- Some concepts to implement: microservices, choreographed saga pattern (async messaging), outbox pattern, idempotency, optimistic updates etc

## Design

A microservice architecture using synchronous and asynchronous communication. There are two Go microservices (Accounts and Fraud). Each have their own database and have access to the message broker for publishing and subscribing to messages. 

The high-level architecture, payment flow and data design are documented below.

![Diagram of system design diagram](./docs/system-design-diagram.svg)

### Tech stack choice:
* **Go**: fast, efficient, simple and development time is quick.
* **Gin**: chosen web framework - it offers some features that make building APIs easier over just using the standard library.
* **RabbitMQ**: it's popular, met my requirements (pub/sub with persistence) and had great documentation.
* **River**: I required a relay for implementing the outbox pattern for reliability, this library was capable of that without me having to build one from scratch, and it had great support for the database I was using straight out of the box.
* **PostgreSQL**: atomicity support - for a banking ledger accuracy is important so I wanted to use a relation database that was ACID compliant.
* **SQLC**: I wanted fine grained control of my SQL statements so I avoided an ORM. But at the same time I didn't want to write everything manually as it could be slow and tedious, SQLC seemed to be the perfect inbetween. It generates Go code based off of my raw SQL statements, speeding up development.
* **Vue.js/tailwind CSS**: this was used to build the frontend demo: for visualising and testing the payment process end-to end. Vue allowed me to build the UI using components in a reactive manner - this enabled me to easily reuse code and build quickly. Tailwind CSS offers pre-built CSS classes which I used without having to write them from scratch. Meaning I had more time to focus on the business logic of the project.

## Running the ledger

All services run inside docker containers using docker-compose. The only prerequisite should be that you have docker installed. Then you can simply run:

```bash
docker compose build
```

You can then visit the frontend here:

```bash
http://localhost:8080
```

Note: some of the database migration/seeding at the moment needs improvements - you may run into issues when rerunning them twice. This is on the list for me to fix, the current implementation is a bit niave. You can simply rebuild if you run into any issues.

## Testing

To run unit tests you can run the following

```bash
cd accounts
go test ./...
```

At the moment the unit tests are lacking some coverage. 

In addition to unit testing, the RabbitMQ dashboard is great for testing messages with different payloads. I found it very valuable for verifying that my idempotency keys were working correctly by replaying the same messages to ensure payments were only processed once. 

## Development

While developing you can run the stack using `watch` which will make the containers automatically rebuild on code changes.

```bash
docker compose watch
```

The RabbitMQ management dashboard is good for troubleshooting queues/messsages. You can even fire adhoc messages on to the queue for testing. You can access it here: `http://localhost:8080/queue`

Default credentials

```
Username: guest
Password: guest
```

SQLC was used for generate Go code for my SQL queries. The process is as follows; write the raw SQL inside the queries `query.sql` and migration files then run the follow command to generate the code.

```bash
sqlc generate
```

You can configure where the migration and query files live inside `sqlc.yaml`.

## Future additions

### Known Issues
- There's no authentication or authorisation. However, this was intentional as the main focus of the project was to build the base features and focus on the various system design techniques. It could be a future addition though.
- Currency GBP is assumed
- Lack of error handling on demo frontend - users should be shown user-friendly errors when something goes wrong.
- Current migration process is limited and can only be run once. In development this may result in having to rebuild the container.
- In development Nginx may restart if a downstream service is down - obviously would not be suitable for the real world but not as a big deal for dev.

### Todo: next features/changes
- Replace current migration program with golang-migrate to enhance development experience
- Add dead letter queues for failed messages - ability to investigate and then replay
- Add delayed retry mechanisms (exponential backoff) and look into circuit breakers?
- Enhance status response codes and errors from unhappy paths
- Utilise Goroutines where appropriate - add worker pool pattern for job queues. They have single worker at the moment.
- Increase test coverage
- Add missing some timeouts via context where appropriate
- Update services to use gRPC (internal comms only)
- Demo frontend: handle errors in friendly manner
- Additional business logic for payments e.g. take into account overdrafts etc.

## Useful docs/reading material
* Dealing with currency: https://cardinalby.github.io/blog/post/best-practices/storing-currency-values-data-types/#1-integer-number-of-minor-units
* RabbitMQ intro: https://www.rabbitmq.com/tutorials/tutorial-one-go
* RabbitMQ pub/sub and exchanges: https://www.rabbitmq.com/tutorials/tutorial-three-go
* River docs: https://riverqueue.com/docs
* Postgres constraints: https://www.postgresql.org/docs/current/ddl-constraints.html
* Golang DB access: https://www.alexedwards.net/blog/organising-database-access
* Visualise go dependencies: https://github.com/kisielk/godepgraph
