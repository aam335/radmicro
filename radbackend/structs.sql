create table users (
    id integer not null primary key,
    username varchar(15),
    attrname varchar(25),
    attrvalue varchar(250)
);

create table acc(
    id integer not null primary key,
    username varchar(15),
    sessionid varchar(100),
    accstart timestamp,
    accstop timestamp,
    inb integer,
    outb integer
);

insert into users(username,attrname,attrvalue)
values
    ("00:00:00:00:00:00", "Auth-Type", "Accept"),
	( "00:00:00:00:00:00",  "Acct-Interim-Interval", "600"),
	( "00:00:00:00:00:00",  "Framed-IP-Address", "192.168.99.100"),
	( "00:00:00:00:00:00",  "Framed-IP-Netmask", "255.255.255.0"),
	( "00:00:00:00:00:00",  "Session-Timeout", "600"),
	( "ut",  "attr1", "val1"),
	( "ut",  "attr2", "val2");