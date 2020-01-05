create table users (
    id integer not null primary key,
    name varchar(15),
    attrname varchar(25),
    attrvalue varchar(250)
);

create table acc(
    id integer not null primary key,
    name varchar(15),
    sessionid varchar(100),
    start timestamp,
    stop timestamp,
    inb integer,
    outb integer
);