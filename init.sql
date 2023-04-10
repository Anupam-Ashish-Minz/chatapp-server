create table users (
	id integer primary key autoincrement,
	name text not null,
	email text not null unique,
	password text not null
);

-- insert into table zaoeuotaeuhs (aoeu) values (25);

insert into users (name, email, password) values ('foo', 'foo@mail.com', 'pass');
insert into users (name, email, password) values ('bar', 'bar@mail.com', 'pass');
insert into users (name, email, password) values ('baz', 'baz@mail.com', 'pass');

create table chatrooms (
	id integer primary key autoincrement,
	name text not null
);

insert into chatrooms (name) values ('dragon');
insert into chatrooms (name) values ('terminator');
insert into chatrooms (name) values ('lightyear');

create table members (
	user_id integer,
	chatroom_id integer,
	-- 0 is admin, 1 is moderator, 2 is ordinary user
	privilage integer not null,
	foreign key (user_id) references users(id),
	foreign key (chatroom_id) references chatrooms(id)
);

insert into members (user_id, chatroom_id, privilage) values (1, 1, 0);
insert into members (user_id, chatroom_id, privilage) values (2, 1, 1);
insert into members (user_id, chatroom_id, privilage) values (3, 1, 2);
insert into members (user_id, chatroom_id, privilage) values (1, 2, 0);
insert into members (user_id, chatroom_id, privilage) values (2, 2, 2);
insert into members (user_id, chatroom_id, privilage) values (3, 2, 2);
insert into members (user_id, chatroom_id, privilage) values (1, 3, 2);
insert into members (user_id, chatroom_id, privilage) values (2, 3, 2);
insert into members (user_id, chatroom_id, privilage) values (3, 3, 0);

create table messages (
	id integer primary key autoincrement,
	time text not null,
	body text not null,
	user_id integer,
	chatroom_id integer,
	foreign key (user_id) references users(id),
	foreign key (chatroom_id) references chatrooms(id)
);

insert into messages (time, body, user_id, chatroom_id) values ('2023-04-10T04:46:30+00:00', 'hello world', 1, 1);
insert into messages (time, body, user_id, chatroom_id) values ('2023-04-10T04:47:22+00:00', 'lorem ipsum', 2, 1);
insert into messages (time, body, user_id, chatroom_id) values ('2023-04-10T04:48:43+00:00', 'this is a text message', 3, 1);
insert into messages (time, body, user_id, chatroom_id) values ('2023-04-10T04:48:43+00:00', 'first', 3, 2);
