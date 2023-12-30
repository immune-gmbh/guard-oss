SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: next_uuid_seq(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.next_uuid_seq(seq text) RETURNS uuid
    LANGUAGE sql
    AS $$ SELECT (lpad(to_hex((nextval(seq) >> 16) & x'ffff'::bigint), 4, '0') || substr(gen_random_uuid()::TEXT, 5))::uuid $$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: addresses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.addresses (
    public_id uuid DEFAULT public.next_uuid_seq('uuid_comb_sequence'::text) NOT NULL,
    street_and_number character varying,
    city character varying,
    postal_code character varying,
    country character varying,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: ar_internal_metadata; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ar_internal_metadata (
    key character varying NOT NULL,
    value character varying,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: authentications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.authentications (
    id bigint NOT NULL,
    user_id uuid NOT NULL,
    provider character varying NOT NULL,
    uid character varying NOT NULL,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: authentications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.authentications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: authentications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.authentications_id_seq OWNED BY public.authentications.id;


--
-- Name: invoices; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoices (
    public_id uuid DEFAULT public.next_uuid_seq('uuid_comb_sequence'::text) NOT NULL,
    subscription_id uuid,
    stripe_invoice_id character varying,
    stripe_invoice_number character varying,
    stripe_pdf_url character varying,
    finalized_at timestamp without time zone,
    paid_at timestamp without time zone,
    marked_uncollectible_at timestamp without time zone,
    voided_at timestamp without time zone,
    tax_rate double precision,
    subtotal integer,
    total integer,
    status character varying,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: memberships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.memberships (
    public_id uuid DEFAULT public.next_uuid_seq('uuid_comb_sequence'::text) NOT NULL,
    user_id uuid,
    organisation_id uuid,
    role integer DEFAULT 2 NOT NULL,
    status integer DEFAULT 0 NOT NULL,
    jwt_token_key character varying,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL,
    notify_device_update boolean DEFAULT false,
    notify_invoice boolean DEFAULT false
);


--
-- Name: organisations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organisations (
    public_id uuid DEFAULT public.next_uuid_seq('uuid_comb_sequence'::text) NOT NULL,
    name character varying,
    vat_number character varying,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL,
    address_id uuid,
    stripe_customer_id character varying,
    contact character varying,
    invoice_email_address character varying,
    invoice_name character varying,
    splunk_enabled boolean,
    splunk_event_collector_url character varying,
    splunk_authentication_token character varying,
    syslog_enabled boolean,
    syslog_hostname_or_address character varying,
    syslog_udp_port character varying,
    splunk_accept_all_server_certificates boolean DEFAULT true,
    status integer DEFAULT 0,
    freeloader boolean DEFAULT false NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subscriptions (
    public_id uuid DEFAULT public.next_uuid_seq('uuid_comb_sequence'::text) NOT NULL,
    stripe_subscription_id character varying,
    stripe_subscription_item_id character varying,
    status character varying DEFAULT 'created'::character varying NOT NULL,
    current_devices_amount integer DEFAULT 0 NOT NULL,
    new_devices_amount integer,
    notify_invoices boolean DEFAULT false NOT NULL,
    notify_device_updates boolean DEFAULT false NOT NULL,
    tax_rate double precision,
    period_start date,
    period_end date,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL,
    address_id uuid,
    organisation_id uuid
);


--
-- Name: usage_records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.usage_records (
    id bigint NOT NULL,
    subscription_id uuid NOT NULL,
    amount integer NOT NULL,
    date date NOT NULL,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: usage_records_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.usage_records_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: usage_records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.usage_records_id_seq OWNED BY public.usage_records.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    public_id uuid DEFAULT public.next_uuid_seq('uuid_comb_sequence'::text) NOT NULL,
    name character varying,
    email character varying NOT NULL,
    crypted_password character varying,
    salt character varying,
    role integer DEFAULT 1,
    invited boolean DEFAULT false,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL,
    activation_state character varying,
    activation_token character varying,
    activation_token_expires_at timestamp without time zone,
    reset_password_token character varying,
    reset_password_token_expires_at timestamp without time zone,
    reset_password_email_sent_at timestamp without time zone,
    access_count_to_reset_password_page integer DEFAULT 0,
    address_id uuid,
    login_token character varying,
    has_seen_intro boolean DEFAULT false,
    login_status integer DEFAULT 0 NOT NULL
);


--
-- Name: uuid_comb_sequence; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.uuid_comb_sequence
    START WITH 70000
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: authentications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.authentications ALTER COLUMN id SET DEFAULT nextval('public.authentications_id_seq'::regclass);


--
-- Name: usage_records id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.usage_records ALTER COLUMN id SET DEFAULT nextval('public.usage_records_id_seq'::regclass);


--
-- Name: addresses addresses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.addresses
    ADD CONSTRAINT addresses_pkey PRIMARY KEY (public_id);


--
-- Name: ar_internal_metadata ar_internal_metadata_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ar_internal_metadata
    ADD CONSTRAINT ar_internal_metadata_pkey PRIMARY KEY (key);


--
-- Name: authentications authentications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.authentications
    ADD CONSTRAINT authentications_pkey PRIMARY KEY (id);


--
-- Name: invoices invoices_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_pkey PRIMARY KEY (public_id);


--
-- Name: memberships memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.memberships
    ADD CONSTRAINT memberships_pkey PRIMARY KEY (public_id);


--
-- Name: organisations organisations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organisations
    ADD CONSTRAINT organisations_pkey PRIMARY KEY (public_id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (public_id);


--
-- Name: usage_records usage_records_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.usage_records
    ADD CONSTRAINT usage_records_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (public_id);


--
-- Name: index_authentications_on_provider_and_uid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_authentications_on_provider_and_uid ON public.authentications USING btree (provider, uid);


--
-- Name: index_invoices_on_subscription_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_invoices_on_subscription_id ON public.invoices USING btree (subscription_id);


--
-- Name: index_memberships_on_organisation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_memberships_on_organisation_id ON public.memberships USING btree (organisation_id);


--
-- Name: index_memberships_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_memberships_on_user_id ON public.memberships USING btree (user_id);


--
-- Name: index_memberships_on_user_id_and_organisation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_memberships_on_user_id_and_organisation_id ON public.memberships USING btree (user_id, organisation_id);


--
-- Name: index_organisations_on_address_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_organisations_on_address_id ON public.organisations USING btree (address_id);


--
-- Name: index_subscriptions_on_address_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_subscriptions_on_address_id ON public.subscriptions USING btree (address_id);


--
-- Name: index_subscriptions_on_organisation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_subscriptions_on_organisation_id ON public.subscriptions USING btree (organisation_id);


--
-- Name: index_usage_records_on_subscription_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_usage_records_on_subscription_id ON public.usage_records USING btree (subscription_id);


--
-- Name: index_users_on_activation_token; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_activation_token ON public.users USING btree (activation_token);


--
-- Name: index_users_on_address_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_address_id ON public.users USING btree (address_id);


--
-- Name: index_users_on_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_users_on_email ON public.users USING btree (email);


--
-- Name: index_users_on_reset_password_token; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_reset_password_token ON public.users USING btree (reset_password_token);


--
-- PostgreSQL database dump complete
--

SET search_path TO "$user", public;

INSERT INTO "schema_migrations" (version) VALUES
('20210618075020'),
('20210618075219'),
('20210618075411'),
('20210618092857'),
('20210623081901'),
('20210623085045'),
('20210624083831'),
('20210628075428'),
('20210628080208'),
('20210629075125'),
('20210630205139'),
('20210630205636'),
('20210707071831'),
('20210714150552'),
('20210721081648'),
('20210805185936'),
('20210806153942'),
('20210809071652'),
('20210810074820'),
('20210903095527'),
('20210906150514'),
('20210909123019'),
('20210920081346'),
('20210929223442'),
('20211001090726'),
('20211010085430');


