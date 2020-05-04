-- INSERT INTO albums (artists, id, name) VALUES ('{pedro}', 14, 'The End');

SELECT * from profiles;

UPDATE profiles SET invitations=array_append(profiles.invitations, cast(albums.id as text)) FROM albums WHERE albums.id=14 AND profiles.userid=2 AND NOT(cast(14 as text)=ANY(profiles.invitations));

-- UPDATE albums SET artists=array_remove(artists, profiles.name) FROM profiles WHERE albums.id=14 AND profiles.userid=2 RETURNING albums.artists;

SELECT * FROM albums;

