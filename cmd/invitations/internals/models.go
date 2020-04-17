package internals

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const schema = `
	CREATE TABLE profiles (
		name text,
		userid integer,
		invitations text[]
	)
`

// Profile has the userID and the username of the profile
type Profile struct {
	Name   string `db:"name"`
	UserID uint32 `db:"userid"`
	// The invitations of this profile to other albums (ID)
	Invitations pq.StringArray `db:"invitations"`
	AlbumName   string         `db:"albumname"`
}

const schemaAlbum = `
	CREATE TABLE albums (
		id integer,
		artists text[],
		name text
	)
`

// Album has the id of the album and the artists collaborating in it
type Album struct {
	ID                   uint32   `db:"id"`
	Name                 string   `db:"name"`
	ArtistsCollaborating []string `db:"artists"`
}

type repo struct {
	db *sqlx.DB
}

// Repository is the database interface
type Repository interface {
	CreateProfile(id uint32, name string) *Profile
	NewInvitation(id uint32, albumID uint32) error
	CreateAlbum(id uint32, artists []string) *Album
	GetInvitations(id uint32) []Invitation
	DeleteInvitation(id uint32, albumID uint32) ([]string, error)
	AcceptInvitation(id uint32, albumID uint32) ([]string, error)
}

// Initialize inits the database
func Initialize(db *sqlx.DB) Repository {
	rep := &repo{db: db}

	tx := db.MustBegin()

	tx.Exec(schema)
	tx.Commit()
	tx = db.MustBegin()
	tx.Exec(schemaAlbum)
	tx.Commit()
	return rep
}

// CreateAlbum creates a album in the database
func (r *repo) CreateAlbum(id uint32, artists []string) *Album {
	tx := r.db.MustBegin()
	row := tx.QueryRowx("INSERT INTO albums (id, artists) VALUES ($1, $2) RETURNING id", id, pq.StringArray(artists))
	if row.Err() != nil {
		fmt.Println("Error: " + row.Err().Error())
		return nil
	}
	tx.Commit()
	return &Album{ID: id}
}

func (r *repo) CreateProfile(id uint32, name string) *Profile {
	tx := r.db.MustBegin()
	row := tx.QueryRowx("INSERT INTO profiles (userid, name, invitations) VALUES ($1, $2, $3) RETURNING userid", id, name, pq.StringArray([]string{}))
	if row.Err() != nil {
		fmt.Println("Error: " + row.Err().Error())
		return nil
	}
	tx.Commit()
	return &Profile{UserID: id, Name: name}
}

// Invitation inv struct
type Invitation struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetInvitations returns all the invitations to the profile
func (r *repo) GetInvitations(id uint32) []Invitation {
	tx := r.db.MustBegin()
	var profile []Profile
	err := tx.Select(&profile, "SELECT profiles.invitations::text[], userid, albums.name as albumname FROM profiles, albums WHERE userid=$1 AND cast(albums.id as text)=ANY(profiles.invitations)", id)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return []Invitation{}
	}
	tx.Commit()
	if len(profile) <= 0 {
		return []Invitation{}
	}

	invs := make([]Invitation, len(profile[0].Invitations))
	for i := range profile[0].Invitations {
		invs[i] = Invitation{Name: profile[0].AlbumName, ID: profile[0].Invitations[i]}
	}

	return invs
}

func (r *repo) NewInvitation(id uint32, albumID uint32) error {
	tx := r.db.MustBegin()

	row, err := tx.Exec("UPDATE profiles SET invitations=array_append(invitations, cast(albums.id as text)) FROM albums WHERE albums.id=$1 AND profiles.id=$2 AND NOT(cast($1 as text)=ANY(profiles.invitations)) AND ", id, albumID)
	if err != nil {
		return err
	}
	n, err := row.RowsAffected()
	if err != nil {
		return err
	}
	tx.Commit()
	if n == 0 {
		return fmt.Errorf("error: album does not exist, the profile already has the invitation or the profile is already collaborating in the album")
	}
	return nil
}

func (r *repo) DeleteInvitation(id uint32, albumID uint32) ([]string, error) {
	tx := r.db.MustBegin()
	var invitations pq.StringArray
	row := tx.QueryRowx("UPDATE profiles SET invitations=array_remove(invitations, cast(albums.id as text)) FROM albums WHERE albums.id=$1 AND profiles.userid=$2 AND cast(albums.id as text)=ANY(profiles.invitations) RETURNING invitations", albumID, id).Scan(&invitations)
	errStr := ""
	if row != nil {
		errStr = row.Error()
	}
	var err error
	if errStr != "" {
		err = fmt.Errorf(errStr)
	}
	if err != nil {
		return nil, err
	}
	tx.Commit()
	invParsed := []string(invitations)
	return invParsed, nil
}

func (r *repo) AcceptInvitation(id uint32, albumID uint32) ([]string, error) {
	if _, err := r.DeleteInvitation(id, albumID); err != nil {
		return nil, err
	}
	tx := r.db.MustBegin()
	var artists pq.StringArray
	row := tx.QueryRowx("UPDATE albums SET artists=array_append(artists, profiles.name) FROM profiles WHERE albums.id=$1 AND profiles.userid=$2 RETURNING albums.artists", albumID, id).Scan(&artists)
	if row != nil && row.Error() != "" {
		return nil, fmt.Errorf(row.Error())
	}
	tx.Commit()

	return artists, nil
}
