package main

import (
	"context"
	"reflect"
	"testing"
)

func TestPrepared_CUD(t *testing.T) {
	db, _ := newDb("CUD")

	type fields struct {
		query     string
		arguments []string
		cacheable bool
	}
	type args struct {
		ctx   context.Context
		attrs map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "insert", fields: fields{query: "insert into users(username) values(:1)", arguments: []string{"user"}},
			args: args{attrs: map[string]string{"user": "vasya"}}},
		{name: "update", fields: fields{query: "update users set attrName=:1, attrValue=:2 where username=:3",
			arguments: []string{"user", "user", "user"}},
			args: args{attrs: map[string]string{"user": "vasya"}}},
		{name: "delete", fields: fields{query: "delete from users where username=:1",
			arguments: []string{"user"}},
			args: args{attrs: map[string]string{"user": "vasya"}}},
		{name: "wrong field count", fields: fields{query: "insert into users(username) values(:1)"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Prepared{
				arguments: tt.fields.arguments,
				cacheable: tt.fields.cacheable,
			}
			var err error
			if q.stmt, err = db.Prepare(tt.fields.query); err != nil {
				t.Fatalf("Prepared.CUD() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.args.ctx = context.Background()
			if err = q.CUD(tt.args.ctx, tt.args.attrs); (err != nil) != tt.wantErr {
				t.Errorf("Prepared.CUD() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		// dumpDB(db, "users")
	}

}

func TestPrepared_R(t *testing.T) {
	db, _ := newDb("R")
	type fields struct {
		query     string
		arguments []string
		cacheable bool
	}
	type args struct {
		ctx   context.Context
		attrs map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]string
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "ut1", fields: fields{query: "select attrname,attrvalue from users where username=:1",
			arguments: []string{"user"}},
			args: args{attrs: map[string]string{"user": "ut"}},
			want: map[string]string{"attr1": "val1", "attr2": "val2"},
		},
		{name: "wrong field count", fields: fields{query: "select attrname,attrvalue from users where username=:1",
			arguments: []string{}},
			args:    args{attrs: map[string]string{"user": "ut"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Prepared{
				arguments: tt.fields.arguments,
				cacheable: tt.fields.cacheable,
			}
			var err error
			if q.stmt, err = db.Prepare(tt.fields.query); err != nil {
				t.Fatalf("Prepared.CUD() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := q.R(context.Background(), tt.args.attrs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Prepared.R() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Prepared.R() = %v, want %v", got, tt.want)
			}
		})
	}
}
