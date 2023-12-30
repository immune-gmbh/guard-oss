class UuidCombSequence < ActiveRecord::Migration[6.1]
  def up
    execute <<-SQL
      CREATE SEQUENCE uuid_comb_sequence START WITH 70000;
      CREATE OR REPLACE FUNCTION next_uuid_seq(seq text) RETURNS uuid
      AS $$ SELECT (lpad(to_hex((nextval(seq) >> 16) & x'ffff'::bigint), 4, '0') || substr(gen_random_uuid()::TEXT, 5))::uuid $$
      LANGUAGE SQL
    SQL
  end

  def down
    execute <<-SQL
      DROP SEQUENCE uuid_comb_sequence
      DROP FUNCTION next_uuid_seq(seq text)
    SQL
  end
end
