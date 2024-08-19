package show

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"

	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/maptile"
)

var re_path = regexp.MustCompile(`/(.*)/(\d+)/(\d+)/(\d+).(\w+)$`)

type TileHandlerOptions struct {
	Database   *sql.DB
	Datasource string
}

func NewTileHandler(opts *TileHandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		t, err := getTileForRequest(req)

		if err != nil {
			slog.Error("Failed to get tile for request", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		d, err := getDataForTile(req, opts, t)

		if err != nil {
			slog.Error("Failed to get data for tile", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		slog.Info("DATA", "data", d)
	}

	return http.HandlerFunc(fn), nil
}

func getDataForTile(req *http.Request, opts *TileHandlerOptions, t *maptile.Tile) (any, error) {

	ctx := req.Context()

	bound := t.Bound()
	poly := bound.ToPolygon()

	enc_poly, err := wkb.MarshalToHex(poly, wkb.DefaultByteOrder)

	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf(`SELECT "wof:id","wof:name", ST_GeomFromWkb(geometry) AS geometry FROM read_parquet("%s") WHERE ST_Intersects(ST_GeomFromWkb(geometry), ST_GeomFromHEXWKB(?))`, opts.Datasource)

	slog.Info(q)
	
	rows, err := opts.Database.QueryContext(ctx, q, string(enc_poly))

	if err != nil {
		slog.Error("Failed to query database", "error", err, "query", q, "geom", enc_poly)
		return nil, fmt.Errorf("Failed to query database, %w", err)
	}

	defer rows.Close()

	for rows.Next() {

		var id int64
		var name string
		var geometry string

		err := rows.Scan(&id, &name, &geometry)

		if err != nil {
			slog.Error("Failed to scan row", "error", err)
			return nil, fmt.Errorf("Failed to scan row, %w", err)
		}

		slog.Info("ROW", "id", id, "name", name, "geometry", geometry)
	}

	err = rows.Err()

	if err != nil {
		// return nil fmt.Errorf("There was a problem scanning rows, %w", err)
		return nil, err
	}

	return nil, nil
}

func getTileForRequest(req *http.Request) (*maptile.Tile, error) {

	path := req.URL.Path

	if !re_path.MatchString(path) {
		return nil, fmt.Errorf("Invalid path")
	}

	m := re_path.FindStringSubmatch(path)

	z, err := strconv.Atoi(m[2])

	if err != nil {
		return nil, err
	}

	x, err := strconv.Atoi(m[3])

	if err != nil {
		return nil, err
	}

	y, err := strconv.Atoi(m[4])

	if err != nil {
		return nil, err
	}

	zm := maptile.Zoom(uint32(z))

	t := &maptile.Tile{
		Z: zm,
		X: uint32(x),
		Y: uint32(y),
	}

	return t, nil
}
