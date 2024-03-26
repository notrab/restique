## Todos

- [] Decide on URL structure

- Use some ideas from postgREST and https://jsonapi.org

### List

- Implement an endpont (e.g. `GET /{tableName}`) that returns a list of all objects
- Should have a default limit
- Allow query string for offset/limit values

### Get

- [ ] Implement an endpoint (e.g. `GET /{tableName}/{primaryKeyValue}`) that returns a single object by `id`
- [ ] Uses `id` as primary key by default
  - [ ] Allows override of primary key
  - [ ] Allow composite PK?
  - [ ] Allow users to include related objects via query string, OR new endpoint entirely (e.g. `GET /posts/1/comments`)

### Create

- Implement an endpoint (e.g. `POST /{tableName}`) that receives JSON and forwards it to the database
- Must return the newly created object (including any DEFAULT, AUTOINCREMENTING, etc. fields)

### Update

- [ ] Implement an endpoint (e.g. `/{tableName}/{primaryKey}) that receives a `PUT`(and or`PATCH`) to replace or patch existing items.
- [ ] Return updated object as JSON

### Delete

- Implement an endpoint (e.g. `DELETE /{tableName}/{primary}`) that removes a row by primary key
- Return deleted object as JSON?

### Includes

- [ ] Support query string to include relational fields

## Todo

- All
- Read
- Create
- Update
- Delete
- Include
