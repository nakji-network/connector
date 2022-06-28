// Code generated by SQLBoiler 3.5.0-gct (https://github.com/thrasher-corp/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package postgres

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries"
	"github.com/thrasher-corp/sqlboiler/randomize"
	"github.com/thrasher-corp/sqlboiler/strmangle"
)

var (
	// Relationships sometimes use the reflection helper queries.Equal/queries.Assign
	// so force a package dependency in case they don't.
	_ = queries.Equal
)

func testCandles(t *testing.T) {
	t.Parallel()

	query := Candles()

	if query.Query == nil {
		t.Error("expected a query, got nothing")
	}
}

func testCandlesDelete(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := o.Delete(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Candles().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testCandlesQueryDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := Candles().DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Candles().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testCandlesSliceDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := CandleSlice{o}

	if rowsAff, err := slice.DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Candles().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testCandlesExists(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	e, err := CandleExists(ctx, tx, o.ID)
	if err != nil {
		t.Errorf("Unable to check if Candle exists: %s", err)
	}
	if !e {
		t.Errorf("Expected CandleExists to return true, but got false.")
	}
}

func testCandlesFind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	candleFound, err := FindCandle(ctx, tx, o.ID)
	if err != nil {
		t.Error(err)
	}

	if candleFound == nil {
		t.Error("want a record, got nil")
	}
}

func testCandlesBind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = Candles().Bind(ctx, tx, o); err != nil {
		t.Error(err)
	}
}

func testCandlesOne(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if x, err := Candles().One(ctx, tx); err != nil {
		t.Error(err)
	} else if x == nil {
		t.Error("expected to get a non nil record")
	}
}

func testCandlesAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	candleOne := &Candle{}
	candleTwo := &Candle{}
	if err = randomize.Struct(seed, candleOne, candleDBTypes, false, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}
	if err = randomize.Struct(seed, candleTwo, candleDBTypes, false, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = candleOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = candleTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := Candles().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 2 {
		t.Error("want 2 records, got:", len(slice))
	}
}

func testCandlesCount(t *testing.T) {
	t.Parallel()

	var err error
	seed := randomize.NewSeed()
	candleOne := &Candle{}
	candleTwo := &Candle{}
	if err = randomize.Struct(seed, candleOne, candleDBTypes, false, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}
	if err = randomize.Struct(seed, candleTwo, candleDBTypes, false, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = candleOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = candleTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Candles().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 2 {
		t.Error("want 2 records, got:", count)
	}
}

func candleBeforeInsertHook(ctx context.Context, e boil.ContextExecutor, o *Candle) error {
	*o = Candle{}
	return nil
}

func candleAfterInsertHook(ctx context.Context, e boil.ContextExecutor, o *Candle) error {
	*o = Candle{}
	return nil
}

func candleAfterSelectHook(ctx context.Context, e boil.ContextExecutor, o *Candle) error {
	*o = Candle{}
	return nil
}

func candleBeforeUpdateHook(ctx context.Context, e boil.ContextExecutor, o *Candle) error {
	*o = Candle{}
	return nil
}

func candleAfterUpdateHook(ctx context.Context, e boil.ContextExecutor, o *Candle) error {
	*o = Candle{}
	return nil
}

func candleBeforeDeleteHook(ctx context.Context, e boil.ContextExecutor, o *Candle) error {
	*o = Candle{}
	return nil
}

func candleAfterDeleteHook(ctx context.Context, e boil.ContextExecutor, o *Candle) error {
	*o = Candle{}
	return nil
}

func candleBeforeUpsertHook(ctx context.Context, e boil.ContextExecutor, o *Candle) error {
	*o = Candle{}
	return nil
}

func candleAfterUpsertHook(ctx context.Context, e boil.ContextExecutor, o *Candle) error {
	*o = Candle{}
	return nil
}

func testCandlesHooks(t *testing.T) {
	t.Parallel()

	var err error

	ctx := context.Background()
	empty := &Candle{}
	o := &Candle{}

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, o, candleDBTypes, false); err != nil {
		t.Errorf("Unable to randomize Candle object: %s", err)
	}

	AddCandleHook(boil.BeforeInsertHook, candleBeforeInsertHook)
	if err = o.doBeforeInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeInsertHook function to empty object, but got: %#v", o)
	}
	candleBeforeInsertHooks = []CandleHook{}

	AddCandleHook(boil.AfterInsertHook, candleAfterInsertHook)
	if err = o.doAfterInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterInsertHook function to empty object, but got: %#v", o)
	}
	candleAfterInsertHooks = []CandleHook{}

	AddCandleHook(boil.AfterSelectHook, candleAfterSelectHook)
	if err = o.doAfterSelectHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterSelectHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterSelectHook function to empty object, but got: %#v", o)
	}
	candleAfterSelectHooks = []CandleHook{}

	AddCandleHook(boil.BeforeUpdateHook, candleBeforeUpdateHook)
	if err = o.doBeforeUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpdateHook function to empty object, but got: %#v", o)
	}
	candleBeforeUpdateHooks = []CandleHook{}

	AddCandleHook(boil.AfterUpdateHook, candleAfterUpdateHook)
	if err = o.doAfterUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpdateHook function to empty object, but got: %#v", o)
	}
	candleAfterUpdateHooks = []CandleHook{}

	AddCandleHook(boil.BeforeDeleteHook, candleBeforeDeleteHook)
	if err = o.doBeforeDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeDeleteHook function to empty object, but got: %#v", o)
	}
	candleBeforeDeleteHooks = []CandleHook{}

	AddCandleHook(boil.AfterDeleteHook, candleAfterDeleteHook)
	if err = o.doAfterDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterDeleteHook function to empty object, but got: %#v", o)
	}
	candleAfterDeleteHooks = []CandleHook{}

	AddCandleHook(boil.BeforeUpsertHook, candleBeforeUpsertHook)
	if err = o.doBeforeUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpsertHook function to empty object, but got: %#v", o)
	}
	candleBeforeUpsertHooks = []CandleHook{}

	AddCandleHook(boil.AfterUpsertHook, candleAfterUpsertHook)
	if err = o.doAfterUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpsertHook function to empty object, but got: %#v", o)
	}
	candleAfterUpsertHooks = []CandleHook{}
}

func testCandlesInsert(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Candles().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testCandlesInsertWhitelist(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Whitelist(candleColumnsWithoutDefault...)); err != nil {
		t.Error(err)
	}

	count, err := Candles().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testCandleToOneExchangeUsingExchangeName(t *testing.T) {
	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var local Candle
	var foreign Exchange

	seed := randomize.NewSeed()
	if err := randomize.Struct(seed, &local, candleDBTypes, false, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}
	if err := randomize.Struct(seed, &foreign, exchangeDBTypes, false, exchangeColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Exchange struct: %s", err)
	}

	if err := foreign.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	local.ExchangeNameID = foreign.ID
	if err := local.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	check, err := local.ExchangeName().One(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}

	if check.ID != foreign.ID {
		t.Errorf("want: %v, got %v", foreign.ID, check.ID)
	}

	slice := CandleSlice{&local}
	if err = local.L.LoadExchangeName(ctx, tx, false, (*[]*Candle)(&slice), nil); err != nil {
		t.Fatal(err)
	}
	if local.R.ExchangeName == nil {
		t.Error("struct should have been eager loaded")
	}

	local.R.ExchangeName = nil
	if err = local.L.LoadExchangeName(ctx, tx, true, &local, nil); err != nil {
		t.Fatal(err)
	}
	if local.R.ExchangeName == nil {
		t.Error("struct should have been eager loaded")
	}
}

func testCandleToOneSetOpExchangeUsingExchangeName(t *testing.T) {
	var err error

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var a Candle
	var b, c Exchange

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, &a, candleDBTypes, false, strmangle.SetComplement(candlePrimaryKeyColumns, candleColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &b, exchangeDBTypes, false, strmangle.SetComplement(exchangePrimaryKeyColumns, exchangeColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &c, exchangeDBTypes, false, strmangle.SetComplement(exchangePrimaryKeyColumns, exchangeColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}

	if err := a.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = b.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	for i, x := range []*Exchange{&b, &c} {
		err = a.SetExchangeName(ctx, tx, i != 0, x)
		if err != nil {
			t.Fatal(err)
		}

		if a.R.ExchangeName != x {
			t.Error("relationship struct not set to correct value")
		}

		if x.R.ExchangeNameCandles[0] != &a {
			t.Error("failed to append to foreign relationship struct")
		}
		if a.ExchangeNameID != x.ID {
			t.Error("foreign key was wrong value", a.ExchangeNameID)
		}

		zero := reflect.Zero(reflect.TypeOf(a.ExchangeNameID))
		reflect.Indirect(reflect.ValueOf(&a.ExchangeNameID)).Set(zero)

		if err = a.Reload(ctx, tx); err != nil {
			t.Fatal("failed to reload", err)
		}

		if a.ExchangeNameID != x.ID {
			t.Error("foreign key was wrong value", a.ExchangeNameID, x.ID)
		}
	}
}

func testCandlesReload(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = o.Reload(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testCandlesReloadAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := CandleSlice{o}

	if err = slice.ReloadAll(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testCandlesSelect(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := Candles().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 1 {
		t.Error("want one record, got:", len(slice))
	}
}

var (
	candleDBTypes = map[string]string{`ID`: `uuid`, `ExchangeNameID`: `uuid`, `Base`: `character varying`, `Quote`: `character varying`, `Interval`: `bigint`, `Timestamp`: `timestamp with time zone`, `Open`: `double precision`, `High`: `double precision`, `Low`: `double precision`, `Close`: `double precision`, `Volume`: `double precision`, `Asset`: `character varying`}
	_             = bytes.MinRead
)

func testCandlesUpdate(t *testing.T) {
	t.Parallel()

	if 0 == len(candlePrimaryKeyColumns) {
		t.Skip("Skipping table with no primary key columns")
	}
	if len(candleAllColumns) == len(candlePrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Candles().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, candleDBTypes, true, candlePrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	if rowsAff, err := o.Update(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only affect one row but affected", rowsAff)
	}
}

func testCandlesSliceUpdateAll(t *testing.T) {
	t.Parallel()

	if len(candleAllColumns) == len(candlePrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &Candle{}
	if err = randomize.Struct(seed, o, candleDBTypes, true, candleColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Candles().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, candleDBTypes, true, candlePrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	// Remove Primary keys and unique columns from what we plan to update
	var fields []string
	if strmangle.StringSliceMatch(candleAllColumns, candlePrimaryKeyColumns) {
		fields = candleAllColumns
	} else {
		fields = strmangle.SetComplement(
			candleAllColumns,
			candlePrimaryKeyColumns,
		)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	typ := reflect.TypeOf(o).Elem()
	n := typ.NumField()

	updateMap := M{}
	for _, col := range fields {
		for i := 0; i < n; i++ {
			f := typ.Field(i)
			if f.Tag.Get("boil") == col {
				updateMap[col] = value.Field(i).Interface()
			}
		}
	}

	slice := CandleSlice{o}
	if rowsAff, err := slice.UpdateAll(ctx, tx, updateMap); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("wanted one record updated but got", rowsAff)
	}
}

func testCandlesUpsert(t *testing.T) {
	t.Parallel()

	if len(candleAllColumns) == len(candlePrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	// Attempt the INSERT side of an UPSERT
	o := Candle{}
	if err = randomize.Struct(seed, &o, candleDBTypes, true); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Upsert(ctx, tx, false, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert Candle: %s", err)
	}

	count, err := Candles().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}

	// Attempt the UPDATE side of an UPSERT
	if err = randomize.Struct(seed, &o, candleDBTypes, false, candlePrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Candle struct: %s", err)
	}

	if err = o.Upsert(ctx, tx, true, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert Candle: %s", err)
	}

	count, err = Candles().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}
}
