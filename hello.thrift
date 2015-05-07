namespace go hello

struct User {
  1: string firstname,
  2: string lastname,
  3: i32    id,
}

service UserManager {
	User get_user(i32 id),
  bool add_user(User user) ,
}

/** vim: set filetype=java ts=2 sts=2 sw=2 fdm=indent et : */                                                              
