package appraisal

/*
func TestExpire(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping integration test")
	}

	ctx := context.Background()
	tests := map[string]func(*testing.T, *pgxpool.Pool){
		// appraisal.Expire()
		"Expire":        testExpire,
		"ExpireExpired": testExpireExpired,
	}
	pgsqlC := mock.PostgresContainer(t, ctx)
	defer pgsqlC.Terminate(ctx)

	for name, fn := range tests {
		pgsqlC.ResetAndSeed(t, ctx, database.MigrationFiles, seedExpireSql, 16)

		t.Run(name, func(t *testing.T) {
			conn := pgsqlC.Connect(t, ctx)
			defer conn.Close()
			fn(t, conn)
		})
	}
}

func testExpire(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	apprs, orgs, err := Expire(ctx, tx, now)
	if err != nil {
		t.Fatal(err)
	}

	if len(apprs) != 8 || len(orgs) != 8 {
		t.Fatalf("expected 8 expired appraisals, got %d,%d", len(apprs), len(orgs))
	}

	for i := 0; i < 8; i += 1 {
		appr := apprs[i]
		dev := apprs[i].Device
		org := orgs[i]

		switch appr.Id {
		case "102":
			if dev.Id != "100" || org != "ext-id-1" {
				t.Error("appraisal 102")
			}
		case "103":
			if dev.Id != "101" || org != "ext-id-1" {
				t.Error("appraisal 103")
			}
		case "110":
			fallthrough
		case "111":
			fallthrough
		case "112":
			fallthrough
		case "113":
			fallthrough
		case "114":
			fallthrough
		case "115":
			if dev.Id != "110" || org != "ext-id-2" {
				t.Error("appraisal 110-115")
			}
		default:
			t.Errorf("didn't expect appraisl %s", appr.Id)
		}
	}
}

func testExpireExpired(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()
	now := time.Now().UTC()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	_, _, err = Expire(ctx, tx, now)
	if err != nil {
		t.Fatal(err)
	}

	apprs, orgs, err := Expire(ctx, tx, now)
	if err != nil {
		t.Fatal(err)
	}

	if len(apprs) != 0 || len(orgs) != 0 {
		t.Fatalf("expected no expired appraisals, got %d,%d", len(apprs), len(orgs))
	}
}
*/
