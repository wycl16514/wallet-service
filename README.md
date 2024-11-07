# Setup db on mac
 1. install postresSQL by https://postgresapp.com/downloads.html
 2. running the app to open the console
 3. create database by using : create database wallet_service;
 4. create user for given database: create user wallet_user with password '1234';
 5. grant privileges: grant all privileges on database wallet_service to wallet_user;
 6. create users table by: create table users (id serial PRIMARY KEY, name VARCHAR(128) NOT NULL);
 7. create wallets table by: create table wallets (id serial PRIMARY key, user_id int not null references users(id), balance numeric(20, 4) not null default 0, check(balance >= 0));
 8. create transaction table by : create table transactions (id serial primary key, user_id int not null references users(id), type varchar(32) not null, amount numeric(20,4) not null, to_user_id int, created_at timestamp default current_timestamp);
 9. create two user records by:
    insert into users(name) values('bob');
    insert into users(name) values('alice');
    
 10. create two records for wallet by:
    insert into wallets(user_id, balance) values(1, 0);
    insert into wallets(user_id, balance) values(2, 0);
 11.  

# Test db connection:
   1. subfolder of config is package of name config, it is responsible for connecting to postgreSQL then returning a DB object.

# Code explaination
Please check my video explaination: https://youtu.be/abYRo1A4AaI

      
