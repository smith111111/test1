#用户对应push_token表
create table push_tokens(
	userid int(11) not null comment '用户ID',
	appid tinyint(1) not null default 0 comment '应用标识,1:OTC',
	device_token varchar(64) not null default '' comment '推送token',
	platform tinyint(1) not null default 0 comment '所属平台,1:android, 2:ios',
	login_status tinyint(1) not null default 0 comment '登录状态，-1：退出。1登录',
	push_type tinyint(1) not null default 0 comment '推送系统类型, 0:友盟推送, 1:极光推送, 2:小米推送',
	update_time int(11) not null default 0 comment '更新时间',
	unique index uk(userid,appid),
	index dt(device_token)
);

#应用的秘钥管理表
create table push_app_keys(
	appid tinyint(1) not null default 0 comment '应用标识, 1:OTC',
	platform tinyint(1) not null default 0 comment '所属平台, 1:android, 2:ios, 3:h5',
	push_type tinyint(1) not null default 0 comment '推送系统类型, 0:友盟推送, 1:极光推送, 2:小米推送, 2:WebSocket推送',
	appkey varchar(64) not null default '' comment 'AppKey',
	secret varchar(64) not null default '' comment '应用标识对的秘钥',
	package_name varchar(64) not null default '' comment '客户端包名',
	unique index uk(appid,platform,push_type)
);

#插入秘钥数据
insert into push_app_keys(appid,platform,push_type,appkey,secret,package_name) values(1,1,1,"ba963f2cb073e159d4eca529","25a1d591446fe66604a84572","");
insert into push_app_keys(appid,platform,push_type,appkey,secret,package_name) values(1,2,1,"ba963f2cb073e159d4eca529","25a1d591446fe66604a84572","");
insert into push_app_keys(appid,platform,push_type,appkey,secret,package_name) values(1,3,3,"","","");

#推送记录表
create table push_msgs (
	msgid bigint not null primary key comment '消息ID',
	msg_type tinyint(1) not null default 0 comment '消息类型',
	userid int(11) not null comment '推送的用户ID',
	appid tinyint(1) not null default 1 comment '应用标识, 1:OTC',
	platform tinyint(1) not null default 0 comment '所属平台, 1:android, 2:ios',
	login_status tinyint(1) not null default 0 comment '推送登录状态，-1：退出。1登录，9:全部',
	title varchar(64) not null default '' comment '推送的标题',
	text varchar(100) not null default '' comment '推送的内容',
	custom text not null  comment '推动自定义数据',
	push_mode tinyint(1) not null comment '推送模式, 1:listcast(列播), 2:broadcast(广播),等',
	push_status varchar(16) not null default '' comment '推送状态',
	error_code varchar(16) not null default '' comment '推送不成功时错误码',
	push_id varchar(32) not null default '' comment '推送消息唯一标识',
	is_del tinyint(1) not null default 0 comment '是否删除, 0:正常, 1:删除',
	insert_time int(11) not null default 0 comment '插入时间',
	update_time int(11) not null default 0 comment '更新时间',
	unique index uk(userid,msgid,msg_type)
);