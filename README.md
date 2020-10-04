# post_views
Solution for a Post View counter.

1. [post-server.go](post-server.go) : HTTP server that serves the requested post.
 
    a. The user is identified by the jwt present in Authorization http header.
    
    b. A jwt can be generated by sending username and password as a `HTTP POST` request to `/login` endpoint
    
        Sample request {"username":"aditapillai@gmail.com","password":"test"}
    
    c. send an `HTTP GET` request to `/posts/{id}` to fetch a post
    
    d. When a user successfully requests for a post, a post_view event is published to the analytics message queue (kafka with topic `post_views`)
  
2. [analytics-page-counts-handler.go](analytics-page-counts-handler.go) : Kafka consumer for the post_views topic

    a. When the consumer gets a message, it parses it and updates the count on the 'counts' db (cassandra with a counter table)
    
    b. The post count and the user who read the post are kept separately in their individual tables namely post_views and post_users. This is because cassandra restricts you to create have a mutable field in a counter table.
    
    c. post_users.users field is a set of users, represented by their IDs as a string
    d. Schema for post_views table
    
      | column | type    | primary key? | indexed? |
      |--------|---------|--------------|----------|
      | postId | text    | true         | true     |
      | counts | counter | false        | false    |
      
    
    e. Schema for post_users table
      
    
      | column | type    | primary key? | indexed? |
      |--------|---------|--------------|----------|
      | postId | text    | true         | true     |
      | users | set\<text\> | false     | true     |
      
  
3. [analytics-server.go](analytics-server.go) : Http server to run analytics queries

    a. `HTTP GET` to `/analytics/posts/{id}/users` will return all the users who have seen the given post
    
    b. `HTTP GET` to `/analytics/users/{id}/posts` will return all the posts (POST IDs) that the user has seen
