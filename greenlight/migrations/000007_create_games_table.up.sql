CREATE TABLE IF NOT EXISTS games (
                                     id bigserial PRIMARY KEY,
                                     created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
                                     runtime integer NOT NULL,
                                     title text NOT NULL,
                                     description TEXT,
                                     genres text[] NOT NULL,
                                     year integer NOT NULL,
                                     version integer NOT NULL ,
                                     size float,
                                     price float
);
BODY='{"title":"Counter-Strike: Global Offensive","year":2012,"runtime":"500", "genres":"shooting","description":"is a multiplayer video game developed by Valve and Hidden Path Entertainment.","size":"12GB","price":"500$","version":"1.38.5.0"}'
