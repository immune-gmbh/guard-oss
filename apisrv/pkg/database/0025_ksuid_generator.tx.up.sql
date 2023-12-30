create extension pgcrypto;
create function v2.next_ksuid_v2(in ref timestamptz(3) = 'now', in payload bytea = '') returns char(27) as $$
declare
  epoch_ref constant bigint := 1400000000;
  epoch_now constant bigint := extract('epoch' from ref);
  ts        constant bigint := epoch_now - epoch_ref;

  alphabet  constant text := '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';
  digit     int;
  id        numeric(66)     := 0;
  -- string of length 27: log_62(2^8*20) = 26.87
  ksuid     text            := '';
begin
  if length(payload) <> 16 then
    payload := gen_random_bytes(16);
  end if;

  -- ts (4 bytes) || payload (16 bytes)
  payload := decode(lpad(to_hex(ts), 8, '0'), 'hex') || payload;
  for i in 0..(16+4-1) loop
    id := (id * 256) + get_byte(payload, i);
  end loop;

  for i in 1..27 loop
    digit := mod(id, 62) + 1;
    id := div(id, 62);
    ksuid := substring(alphabet from digit for 1) || ksuid;
  end loop;

  return ksuid;
end;
$$ language plpgsql;

create domain v2.ksuid_ref as char(27) 
  check (value similar to '[a-zA-Z0-9]{27}');

create domain v2.ksuid as char(27) 
  default v2.next_ksuid_v2()
  check (value similar to '[a-zA-Z0-9]{27}');
