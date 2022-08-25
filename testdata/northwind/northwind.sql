--
-- PostgreSQL database dump
--

-- Dumped from database version 14.5 (Ubuntu 14.5-0ubuntu0.22.04.1)
-- Dumped by pg_dump version 14.5 (Ubuntu 14.5-0ubuntu0.22.04.1)

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: suppliers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.suppliers (
    supplier_id smallint NOT NULL,
    company_name character varying(40) NOT NULL,
    contact_name character varying(30),
    contact_title character varying(30),
    address character varying(60),
    city character varying(15),
    region character varying(15),
    postal_code character varying(10),
    country character varying(15),
    phone character varying(24),
    fax character varying(24),
    homepage text
);


ALTER TABLE public.suppliers OWNER TO postgres;

--
-- Name: test_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.test_type AS (
	base public.suppliers,
	name text
);


ALTER TYPE public.test_type OWNER TO postgres;

--
-- Name: dynamic_sql(anyelement, text[]); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.dynamic_sql(inp anyelement, VARIADIC params text[]) RETURNS jsonb
    LANGUAGE plpgsql
    AS $_$
declare data jsonb;
declare flds text;
begin
  flds := array_to_string(params,',');
  EXECUTE format('
  select json_agg(aa)
  from (select %s
  FROM   suppliers
  WHERE  supplier_id = $1) aa',
  flds)
   into data
    using 5;
   return data;
end
$_$;


ALTER FUNCTION public.dynamic_sql(inp anyelement, VARIADIC params text[]) OWNER TO postgres;

--
-- Name: get_suppliers(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_suppliers() RETURNS SETOF public.suppliers
    LANGUAGE sql
    AS $$
    SELECT * FROM suppliers limit 10;
$$;


ALTER FUNCTION public.get_suppliers() OWNER TO postgres;

--
-- Name: test_get_rel_rows(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.test_get_rel_rows() RETURNS SETOF record
    LANGUAGE plpgsql
    AS $$
	BEGIN

	END;
$$;


ALTER FUNCTION public.test_get_rel_rows() OWNER TO postgres;

--
-- Name: test_get_rel_rows(text, record); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.test_get_rel_rows(name text, inp record) RETURNS SETOF record
    LANGUAGE plpgsql
    AS $$
    declare rec record;
	BEGIN
		select * from suppliers into rec;
	    return next rec;
	END;
$$;


ALTER FUNCTION public.test_get_rel_rows(name text, inp record) OWNER TO postgres;

--
-- Name: customers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.customers (
    customer_id bpchar NOT NULL,
    company_name character varying(40) NOT NULL,
    contact_name character varying(30),
    contact_title character varying(30),
    address character varying(60),
    city character varying(15),
    region character varying(15),
    postal_code character varying(10),
    country character varying(15),
    phone character varying(24),
    fax character varying(24)
);


ALTER TABLE public.customers OWNER TO postgres;

--
-- Name: test_rs(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.test_rs(inp_customer_id text) RETURNS TABLE(base public.customers, order_id text)
    LANGUAGE plpgsql
    AS $$
	declare
	  test text;
	begin
	  create local temporary table test_tab on commit drop as
	    select c.*, 'aaa'::text as order_id from customers c;
	   for counter in 1..5 loop
		raise notice 'counter: %', counter;
	   end loop;	    
--     test := array_to_string(ARRAY(select order_id from test_tab),',');
--     RAISE NOTICE 'basic message s%', test;
     return query select tt.*, 'aaa'::text from test_tab tt;
	end
$$;


ALTER FUNCTION public.test_rs(inp_customer_id text) OWNER TO postgres;

--
-- Name: categories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.categories (
    category_id smallint NOT NULL,
    category_name character varying(15) NOT NULL,
    description text,
    picture bytea
);


ALTER TABLE public.categories OWNER TO postgres;

--
-- Name: customer_customer_demo; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.customer_customer_demo (
    customer_id bpchar NOT NULL,
    customer_type_id bpchar NOT NULL
);


ALTER TABLE public.customer_customer_demo OWNER TO postgres;

--
-- Name: customer_demographics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.customer_demographics (
    customer_type_id bpchar NOT NULL,
    customer_desc text
);


ALTER TABLE public.customer_demographics OWNER TO postgres;

--
-- Name: employee_territories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.employee_territories (
    employee_id smallint NOT NULL,
    territory_id character varying(20) NOT NULL
);


ALTER TABLE public.employee_territories OWNER TO postgres;

--
-- Name: employees; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.employees (
    employee_id smallint NOT NULL,
    last_name character varying(20) NOT NULL,
    first_name character varying(10) NOT NULL,
    title character varying(30),
    title_of_courtesy character varying(25),
    birth_date date,
    hire_date date,
    address character varying(60),
    city character varying(15),
    region character varying(15),
    postal_code character varying(10),
    country character varying(15),
    home_phone character varying(24),
    extension character varying(4),
    photo bytea,
    notes text,
    reports_to smallint,
    photo_path character varying(255)
);


ALTER TABLE public.employees OWNER TO postgres;

--
-- Name: nc_acl; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_acl (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255) DEFAULT 'db'::character varying,
    tn character varying(255),
    acl text,
    type character varying(255) DEFAULT 'table'::character varying,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_acl OWNER TO postgres;

--
-- Name: nc_acl_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_acl_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_acl_id_seq OWNER TO postgres;

--
-- Name: nc_acl_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_acl_id_seq OWNED BY public.nc_acl.id;


--
-- Name: nc_api_tokens; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_api_tokens (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255),
    description character varying(255),
    permissions text,
    token text,
    expiry character varying(255),
    enabled boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_api_tokens OWNER TO postgres;

--
-- Name: nc_api_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_api_tokens_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_api_tokens_id_seq OWNER TO postgres;

--
-- Name: nc_api_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_api_tokens_id_seq OWNED BY public.nc_api_tokens.id;


--
-- Name: nc_audit; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_audit (
    id integer NOT NULL,
    "user" character varying(255),
    ip character varying(255),
    project_id character varying(255),
    db_alias character varying(255),
    model_name character varying(100),
    model_id character varying(100),
    op_type character varying(255),
    op_sub_type character varying(255),
    status character varying(255),
    description text,
    details text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_audit OWNER TO postgres;

--
-- Name: nc_audit_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_audit_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_audit_id_seq OWNER TO postgres;

--
-- Name: nc_audit_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_audit_id_seq OWNED BY public.nc_audit.id;


--
-- Name: nc_audit_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_audit_v2 (
    id character varying(20) NOT NULL,
    "user" character varying(255),
    ip character varying(255),
    base_id character varying(20),
    project_id character varying(128),
    fk_model_id character varying(20),
    row_id character varying(255),
    op_type character varying(255),
    op_sub_type character varying(255),
    status character varying(255),
    description text,
    details text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_audit_v2 OWNER TO postgres;

--
-- Name: nc_bases_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_bases_v2 (
    id character varying(20) NOT NULL,
    project_id character varying(128),
    alias character varying(255),
    config text,
    meta text,
    is_meta boolean,
    type character varying(255),
    inflection_column character varying(255),
    inflection_table character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_bases_v2 OWNER TO postgres;

--
-- Name: nc_col_formula_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_col_formula_v2 (
    id character varying(20) NOT NULL,
    fk_column_id character varying(20),
    formula text NOT NULL,
    formula_raw text,
    error text,
    deleted boolean,
    "order" real,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_col_formula_v2 OWNER TO postgres;

--
-- Name: nc_col_lookup_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_col_lookup_v2 (
    id character varying(20) NOT NULL,
    fk_column_id character varying(20),
    fk_relation_column_id character varying(20),
    fk_lookup_column_id character varying(20),
    deleted boolean,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_col_lookup_v2 OWNER TO postgres;

--
-- Name: nc_col_relations_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_col_relations_v2 (
    id character varying(20) NOT NULL,
    ref_db_alias character varying(255),
    type character varying(255),
    virtual boolean,
    db_type character varying(255),
    fk_column_id character varying(20),
    fk_related_model_id character varying(20),
    fk_child_column_id character varying(20),
    fk_parent_column_id character varying(20),
    fk_mm_model_id character varying(20),
    fk_mm_child_column_id character varying(20),
    fk_mm_parent_column_id character varying(20),
    ur character varying(255),
    dr character varying(255),
    fk_index_name character varying(255),
    deleted boolean,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_col_relations_v2 OWNER TO postgres;

--
-- Name: nc_col_rollup_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_col_rollup_v2 (
    id character varying(20) NOT NULL,
    fk_column_id character varying(20),
    fk_relation_column_id character varying(20),
    fk_rollup_column_id character varying(20),
    rollup_function character varying(255),
    deleted boolean,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_col_rollup_v2 OWNER TO postgres;

--
-- Name: nc_col_select_options_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_col_select_options_v2 (
    id character varying(20) NOT NULL,
    fk_column_id character varying(20),
    title character varying(255),
    color character varying(255),
    "order" real,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_col_select_options_v2 OWNER TO postgres;

--
-- Name: nc_columns_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_columns_v2 (
    id character varying(20) NOT NULL,
    base_id character varying(20),
    project_id character varying(128),
    fk_model_id character varying(20),
    title character varying(255),
    column_name character varying(255),
    uidt character varying(255),
    dt character varying(255),
    np character varying(255),
    ns character varying(255),
    clen character varying(255),
    cop character varying(255),
    pk boolean,
    pv boolean,
    rqd boolean,
    un boolean,
    ct text,
    ai boolean,
    "unique" boolean,
    cdf text,
    cc text,
    csn character varying(255),
    dtx character varying(255),
    dtxp text,
    dtxs character varying(255),
    au boolean,
    validate text,
    virtual boolean,
    deleted boolean,
    system boolean DEFAULT false,
    "order" real,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_columns_v2 OWNER TO postgres;

--
-- Name: nc_cron; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_cron (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255) DEFAULT 'db'::character varying,
    title character varying(255),
    description character varying(255),
    env character varying(255),
    pattern character varying(255),
    webhook character varying(255),
    timezone character varying(255) DEFAULT 'America/Los_Angeles'::character varying,
    active boolean DEFAULT true,
    cron_handler text,
    payload text,
    headers text,
    retries integer DEFAULT 0,
    retry_interval integer DEFAULT 60000,
    timeout integer DEFAULT 60000,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_cron OWNER TO postgres;

--
-- Name: nc_cron_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_cron_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_cron_id_seq OWNER TO postgres;

--
-- Name: nc_cron_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_cron_id_seq OWNED BY public.nc_cron.id;


--
-- Name: nc_disabled_models_for_role; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_disabled_models_for_role (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(45),
    title character varying(45),
    type character varying(45),
    role character varying(45),
    disabled boolean DEFAULT true,
    tn character varying(255),
    rtn character varying(255),
    cn character varying(255),
    rcn character varying(255),
    relation_type character varying(255),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    parent_model_title character varying(255)
);


ALTER TABLE public.nc_disabled_models_for_role OWNER TO postgres;

--
-- Name: nc_disabled_models_for_role_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_disabled_models_for_role_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_disabled_models_for_role_id_seq OWNER TO postgres;

--
-- Name: nc_disabled_models_for_role_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_disabled_models_for_role_id_seq OWNED BY public.nc_disabled_models_for_role.id;


--
-- Name: nc_disabled_models_for_role_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_disabled_models_for_role_v2 (
    id character varying(20) NOT NULL,
    base_id character varying(20),
    project_id character varying(128),
    fk_view_id character varying(20),
    role character varying(45),
    disabled boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_disabled_models_for_role_v2 OWNER TO postgres;

--
-- Name: nc_evolutions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_evolutions (
    id integer NOT NULL,
    title character varying(255) NOT NULL,
    "titleDown" character varying(255),
    description character varying(255),
    batch integer,
    checksum character varying(255),
    status integer,
    created timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_evolutions OWNER TO postgres;

--
-- Name: nc_evolutions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_evolutions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_evolutions_id_seq OWNER TO postgres;

--
-- Name: nc_evolutions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_evolutions_id_seq OWNED BY public.nc_evolutions.id;


--
-- Name: nc_filter_exp_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_filter_exp_v2 (
    id character varying(20) NOT NULL,
    base_id character varying(20),
    project_id character varying(128),
    fk_view_id character varying(20),
    fk_hook_id character varying(20),
    fk_column_id character varying(20),
    fk_parent_id character varying(20),
    logical_op character varying(255),
    comparison_op character varying(255),
    value character varying(255),
    is_group boolean,
    "order" real,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_filter_exp_v2 OWNER TO postgres;

--
-- Name: nc_form_view_columns_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_form_view_columns_v2 (
    id character varying(20) NOT NULL,
    base_id character varying(20),
    project_id character varying(128),
    fk_view_id character varying(20),
    fk_column_id character varying(20),
    uuid character varying(255),
    label character varying(255),
    help character varying(255),
    description character varying(255),
    required boolean,
    show boolean,
    "order" real,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_form_view_columns_v2 OWNER TO postgres;

--
-- Name: nc_form_view_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_form_view_v2 (
    base_id character varying(20),
    project_id character varying(128),
    fk_view_id character varying(20) NOT NULL,
    heading character varying(255),
    subheading character varying(255),
    success_msg character varying(255),
    redirect_url character varying(255),
    redirect_after_secs character varying(255),
    email character varying(255),
    submit_another_form boolean,
    show_blank_form boolean,
    uuid character varying(255),
    banner_image_url character varying(255),
    logo_url character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_form_view_v2 OWNER TO postgres;

--
-- Name: nc_gallery_view_columns_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_gallery_view_columns_v2 (
    id character varying(20) NOT NULL,
    base_id character varying(20),
    project_id character varying(128),
    fk_view_id character varying(20),
    fk_column_id character varying(20),
    uuid character varying(255),
    label character varying(255),
    help character varying(255),
    show boolean,
    "order" real,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_gallery_view_columns_v2 OWNER TO postgres;

--
-- Name: nc_gallery_view_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_gallery_view_v2 (
    base_id character varying(20),
    project_id character varying(128),
    fk_view_id character varying(20) NOT NULL,
    next_enabled boolean,
    prev_enabled boolean,
    cover_image_idx integer,
    fk_cover_image_col_id character varying(20),
    cover_image character varying(255),
    restrict_types character varying(255),
    restrict_size character varying(255),
    restrict_number character varying(255),
    public boolean,
    dimensions character varying(255),
    responsive_columns character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_gallery_view_v2 OWNER TO postgres;

--
-- Name: nc_grid_view_columns_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_grid_view_columns_v2 (
    id character varying(20) NOT NULL,
    fk_view_id character varying(20),
    fk_column_id character varying(20),
    base_id character varying(20),
    project_id character varying(128),
    uuid character varying(255),
    label character varying(255),
    help character varying(255),
    width character varying(255) DEFAULT '200px'::character varying,
    show boolean,
    "order" real,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_grid_view_columns_v2 OWNER TO postgres;

--
-- Name: nc_grid_view_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_grid_view_v2 (
    fk_view_id character varying(20) NOT NULL,
    base_id character varying(20),
    project_id character varying(128),
    uuid character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_grid_view_v2 OWNER TO postgres;

--
-- Name: nc_hook_logs_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_hook_logs_v2 (
    id character varying(20) NOT NULL,
    base_id character varying(20),
    project_id character varying(128),
    fk_hook_id character varying(20),
    type character varying(255),
    event character varying(255),
    operation character varying(255),
    test_call boolean DEFAULT true,
    payload boolean DEFAULT true,
    conditions text,
    notification text,
    error_code character varying(255),
    error_message character varying(255),
    error text,
    execution_time integer,
    response character varying(255),
    triggered_by character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_hook_logs_v2 OWNER TO postgres;

--
-- Name: nc_hooks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_hooks (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255) DEFAULT 'db'::character varying,
    title character varying(255),
    description character varying(255),
    env character varying(255) DEFAULT 'all'::character varying,
    tn character varying(255),
    type character varying(255),
    event character varying(255),
    operation character varying(255),
    async boolean DEFAULT false,
    payload boolean DEFAULT true,
    url text,
    headers text,
    condition text,
    notification text,
    retries integer DEFAULT 0,
    retry_interval integer DEFAULT 60000,
    timeout integer DEFAULT 60000,
    active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_hooks OWNER TO postgres;

--
-- Name: nc_hooks_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_hooks_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_hooks_id_seq OWNER TO postgres;

--
-- Name: nc_hooks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_hooks_id_seq OWNED BY public.nc_hooks.id;


--
-- Name: nc_hooks_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_hooks_v2 (
    id character varying(20) NOT NULL,
    base_id character varying(20),
    project_id character varying(128),
    fk_model_id character varying(20),
    title character varying(255),
    description character varying(255),
    env character varying(255) DEFAULT 'all'::character varying,
    type character varying(255),
    event character varying(255),
    operation character varying(255),
    async boolean DEFAULT false,
    payload boolean DEFAULT true,
    url text,
    headers text,
    condition boolean DEFAULT false,
    notification text,
    retries integer DEFAULT 0,
    retry_interval integer DEFAULT 60000,
    timeout integer DEFAULT 60000,
    active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_hooks_v2 OWNER TO postgres;

--
-- Name: nc_kanban_view_columns_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_kanban_view_columns_v2 (
    id character varying(20) NOT NULL,
    base_id character varying(20),
    project_id character varying(128),
    fk_view_id character varying(20),
    fk_column_id character varying(20),
    uuid character varying(255),
    label character varying(255),
    help character varying(255),
    show boolean,
    "order" real,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_kanban_view_columns_v2 OWNER TO postgres;

--
-- Name: nc_kanban_view_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_kanban_view_v2 (
    fk_view_id character varying(20) NOT NULL,
    base_id character varying(20),
    project_id character varying(128),
    show boolean,
    "order" real,
    uuid character varying(255),
    title character varying(255),
    public boolean,
    password character varying(255),
    show_all_fields boolean,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_kanban_view_v2 OWNER TO postgres;

--
-- Name: nc_loaders; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_loaders (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255) DEFAULT 'db'::character varying,
    title character varying(255),
    parent character varying(255),
    child character varying(255),
    relation character varying(255),
    resolver character varying(255),
    functions text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_loaders OWNER TO postgres;

--
-- Name: nc_loaders_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_loaders_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_loaders_id_seq OWNER TO postgres;

--
-- Name: nc_loaders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_loaders_id_seq OWNED BY public.nc_loaders.id;


--
-- Name: nc_migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_migrations (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255),
    up text,
    down text,
    title character varying(255) NOT NULL,
    title_down character varying(255),
    description character varying(255),
    batch integer,
    checksum character varying(255),
    status integer,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_migrations OWNER TO postgres;

--
-- Name: nc_migrations_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_migrations_id_seq OWNER TO postgres;

--
-- Name: nc_migrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_migrations_id_seq OWNED BY public.nc_migrations.id;


--
-- Name: nc_models; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_models (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255) DEFAULT 'db'::character varying,
    title character varying(255),
    alias character varying(255),
    type character varying(255) DEFAULT 'table'::character varying,
    meta text,
    schema text,
    schema_previous text,
    services text,
    messages text,
    enabled boolean DEFAULT true,
    parent_model_title character varying(255),
    show_as character varying(255) DEFAULT 'table'::character varying,
    query_params text,
    list_idx integer,
    tags character varying(255),
    pinned boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    mm integer,
    m_to_m_meta text,
    "order" real,
    view_order real
);


ALTER TABLE public.nc_models OWNER TO postgres;

--
-- Name: nc_models_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_models_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_models_id_seq OWNER TO postgres;

--
-- Name: nc_models_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_models_id_seq OWNED BY public.nc_models.id;


--
-- Name: nc_models_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_models_v2 (
    id character varying(20) NOT NULL,
    base_id character varying(20),
    project_id character varying(128),
    table_name character varying(255),
    title character varying(255),
    type character varying(255) DEFAULT 'table'::character varying,
    meta text,
    schema text,
    enabled boolean DEFAULT true,
    mm boolean DEFAULT false,
    tags character varying(255),
    pinned boolean,
    deleted boolean,
    "order" real,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_models_v2 OWNER TO postgres;

--
-- Name: nc_orgs_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_orgs_v2 (
    id character varying(20) NOT NULL,
    title character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_orgs_v2 OWNER TO postgres;

--
-- Name: nc_plugins; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_plugins (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255),
    title character varying(45),
    description text,
    active boolean DEFAULT false,
    rating real,
    version character varying(255),
    docs character varying(255),
    status character varying(255) DEFAULT 'install'::character varying,
    status_details character varying(255),
    logo character varying(255),
    icon character varying(255),
    tags character varying(255),
    category character varying(255),
    input_schema text,
    input text,
    creator character varying(255),
    creator_website character varying(255),
    price character varying(255),
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_plugins OWNER TO postgres;

--
-- Name: nc_plugins_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_plugins_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_plugins_id_seq OWNER TO postgres;

--
-- Name: nc_plugins_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_plugins_id_seq OWNED BY public.nc_plugins.id;


--
-- Name: nc_plugins_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_plugins_v2 (
    id character varying(20) NOT NULL,
    title character varying(45),
    description text,
    active boolean DEFAULT false,
    rating real,
    version character varying(255),
    docs character varying(255),
    status character varying(255) DEFAULT 'install'::character varying,
    status_details character varying(255),
    logo character varying(255),
    icon character varying(255),
    tags character varying(255),
    category character varying(255),
    input_schema text,
    input text,
    creator character varying(255),
    creator_website character varying(255),
    price character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_plugins_v2 OWNER TO postgres;

--
-- Name: nc_project_users_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_project_users_v2 (
    project_id character varying(128),
    fk_user_id character varying(20),
    roles text,
    starred boolean,
    pinned boolean,
    "group" character varying(255),
    color character varying(255),
    "order" real,
    hidden real,
    opened_date timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_project_users_v2 OWNER TO postgres;

--
-- Name: nc_projects; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_projects (
    id character varying(128) NOT NULL,
    title character varying(255),
    status character varying(255),
    description text,
    config text,
    meta text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_projects OWNER TO postgres;

--
-- Name: nc_projects_users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_projects_users (
    project_id character varying(255),
    user_id integer,
    roles text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_projects_users OWNER TO postgres;

--
-- Name: nc_projects_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_projects_v2 (
    id character varying(128) NOT NULL,
    title character varying(255),
    prefix character varying(255),
    status character varying(255),
    description text,
    meta text,
    color character varying(255),
    uuid character varying(255),
    password character varying(255),
    roles character varying(255),
    deleted boolean DEFAULT false,
    is_meta boolean,
    "order" real,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_projects_v2 OWNER TO postgres;

--
-- Name: nc_relations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_relations (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255),
    tn character varying(255),
    rtn character varying(255),
    _tn character varying(255),
    _rtn character varying(255),
    cn character varying(255),
    rcn character varying(255),
    _cn character varying(255),
    _rcn character varying(255),
    referenced_db_alias character varying(255),
    type character varying(255),
    db_type character varying(255),
    ur character varying(255),
    dr character varying(255),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    fkn character varying(255)
);


ALTER TABLE public.nc_relations OWNER TO postgres;

--
-- Name: nc_relations_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_relations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_relations_id_seq OWNER TO postgres;

--
-- Name: nc_relations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_relations_id_seq OWNED BY public.nc_relations.id;


--
-- Name: nc_resolvers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_resolvers (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255) DEFAULT 'db'::character varying,
    title character varying(255),
    resolver text,
    type character varying(255),
    acl text,
    functions text,
    handler_type integer DEFAULT 1,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_resolvers OWNER TO postgres;

--
-- Name: nc_resolvers_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_resolvers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_resolvers_id_seq OWNER TO postgres;

--
-- Name: nc_resolvers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_resolvers_id_seq OWNED BY public.nc_resolvers.id;


--
-- Name: nc_roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_roles (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255) DEFAULT 'db'::character varying,
    title character varying(255),
    type character varying(255) DEFAULT 'CUSTOM'::character varying,
    description character varying(255),
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_roles OWNER TO postgres;

--
-- Name: nc_roles_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_roles_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_roles_id_seq OWNER TO postgres;

--
-- Name: nc_roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_roles_id_seq OWNED BY public.nc_roles.id;


--
-- Name: nc_routes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_routes (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255) DEFAULT 'db'::character varying,
    title character varying(255),
    tn character varying(255),
    tnp character varying(255),
    tnc character varying(255),
    relation_type character varying(255),
    path text,
    type character varying(255),
    handler text,
    acl text,
    "order" integer,
    functions text,
    handler_type integer DEFAULT 1,
    is_custom boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_routes OWNER TO postgres;

--
-- Name: nc_routes_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_routes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_routes_id_seq OWNER TO postgres;

--
-- Name: nc_routes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_routes_id_seq OWNED BY public.nc_routes.id;


--
-- Name: nc_rpc; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_rpc (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255) DEFAULT 'db'::character varying,
    title character varying(255),
    tn character varying(255),
    service text,
    tnp character varying(255),
    tnc character varying(255),
    relation_type character varying(255),
    "order" integer,
    type character varying(255),
    acl text,
    functions text,
    handler_type integer DEFAULT 1,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_rpc OWNER TO postgres;

--
-- Name: nc_rpc_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_rpc_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_rpc_id_seq OWNER TO postgres;

--
-- Name: nc_rpc_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_rpc_id_seq OWNED BY public.nc_rpc.id;


--
-- Name: nc_shared_bases; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_shared_bases (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255),
    roles character varying(255) DEFAULT 'viewer'::character varying,
    shared_base_id character varying(255),
    enabled boolean DEFAULT true,
    password character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_shared_bases OWNER TO postgres;

--
-- Name: nc_shared_bases_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_shared_bases_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_shared_bases_id_seq OWNER TO postgres;

--
-- Name: nc_shared_bases_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_shared_bases_id_seq OWNED BY public.nc_shared_bases.id;


--
-- Name: nc_shared_views; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_shared_views (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255),
    model_name character varying(255),
    meta text,
    query_params text,
    view_id character varying(255),
    show_all_fields boolean,
    allow_copy boolean,
    password character varying(255),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    view_type character varying(255),
    view_name character varying(255)
);


ALTER TABLE public.nc_shared_views OWNER TO postgres;

--
-- Name: nc_shared_views_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_shared_views_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_shared_views_id_seq OWNER TO postgres;

--
-- Name: nc_shared_views_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_shared_views_id_seq OWNED BY public.nc_shared_views.id;


--
-- Name: nc_shared_views_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_shared_views_v2 (
    id character varying(20) NOT NULL,
    fk_view_id character varying(20),
    meta text,
    query_params text,
    view_id character varying(255),
    show_all_fields boolean,
    allow_copy boolean,
    password character varying(255),
    deleted boolean,
    "order" real,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_shared_views_v2 OWNER TO postgres;

--
-- Name: nc_sort_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_sort_v2 (
    id character varying(20) NOT NULL,
    base_id character varying(20),
    project_id character varying(128),
    fk_view_id character varying(20),
    fk_column_id character varying(20),
    direction character varying(255) DEFAULT 'false'::character varying,
    "order" real,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_sort_v2 OWNER TO postgres;

--
-- Name: nc_store; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_store (
    id integer NOT NULL,
    project_id character varying(255),
    db_alias character varying(255) DEFAULT 'db'::character varying,
    key character varying(255),
    value text,
    type character varying(255),
    env character varying(255),
    tag character varying(255),
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.nc_store OWNER TO postgres;

--
-- Name: nc_store_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.nc_store_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.nc_store_id_seq OWNER TO postgres;

--
-- Name: nc_store_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.nc_store_id_seq OWNED BY public.nc_store.id;


--
-- Name: nc_team_users_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_team_users_v2 (
    org_id character varying(20),
    user_id character varying(20),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_team_users_v2 OWNER TO postgres;

--
-- Name: nc_teams_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_teams_v2 (
    id character varying(20) NOT NULL,
    title character varying(255),
    org_id character varying(20),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_teams_v2 OWNER TO postgres;

--
-- Name: nc_users_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_users_v2 (
    id character varying(20) NOT NULL,
    email character varying(255),
    password character varying(255),
    salt character varying(255),
    firstname character varying(255),
    lastname character varying(255),
    username character varying(255),
    refresh_token character varying(255),
    invite_token character varying(255),
    invite_token_expires character varying(255),
    reset_password_expires timestamp with time zone,
    reset_password_token character varying(255),
    email_verification_token character varying(255),
    email_verified boolean,
    roles character varying(255) DEFAULT 'editor'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_users_v2 OWNER TO postgres;

--
-- Name: nc_views_v2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nc_views_v2 (
    id character varying(20) NOT NULL,
    base_id character varying(20),
    project_id character varying(128),
    fk_model_id character varying(20),
    title character varying(255),
    type integer,
    is_default boolean,
    show_system_fields boolean,
    lock_type character varying(255) DEFAULT 'collaborative'::character varying,
    uuid character varying(255),
    password character varying(255),
    show boolean,
    "order" real,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.nc_views_v2 OWNER TO postgres;

--
-- Name: order_details; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.order_details (
    order_id smallint NOT NULL,
    product_id smallint NOT NULL,
    unit_price real NOT NULL,
    quantity smallint NOT NULL,
    discount real NOT NULL
);


ALTER TABLE public.order_details OWNER TO postgres;

--
-- Name: orders; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.orders (
    order_id smallint NOT NULL,
    customer_id bpchar,
    employee_id smallint,
    order_date date,
    required_date date,
    shipped_date date,
    ship_via smallint,
    freight real,
    ship_name character varying(40),
    ship_address character varying(60),
    ship_city character varying(15),
    ship_region character varying(15),
    ship_postal_code character varying(10),
    ship_country character varying(15)
);


ALTER TABLE public.orders OWNER TO postgres;

--
-- Name: products; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.products (
    product_id smallint NOT NULL,
    product_name character varying(40) NOT NULL,
    supplier_id smallint,
    category_id smallint,
    quantity_per_unit character varying(20),
    unit_price real,
    units_in_stock smallint,
    units_on_order smallint,
    reorder_level smallint,
    discontinued integer NOT NULL
);


ALTER TABLE public.products OWNER TO postgres;

--
-- Name: region; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.region (
    region_id smallint NOT NULL,
    region_description bpchar NOT NULL
);


ALTER TABLE public.region OWNER TO postgres;

--
-- Name: shippers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.shippers (
    shipper_id smallint NOT NULL,
    company_name character varying(40) NOT NULL,
    phone character varying(24)
);


ALTER TABLE public.shippers OWNER TO postgres;

--
-- Name: territories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.territories (
    territory_id character varying(20) NOT NULL,
    territory_description bpchar NOT NULL,
    region_id smallint NOT NULL
);


ALTER TABLE public.territories OWNER TO postgres;

--
-- Name: us_states; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.us_states (
    state_id smallint NOT NULL,
    state_name character varying(100),
    state_abbr character varying(2),
    state_region character varying(50)
);


ALTER TABLE public.us_states OWNER TO postgres;

--
-- Name: xc_knex_migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.xc_knex_migrations (
    id integer NOT NULL,
    name character varying(255),
    batch integer,
    migration_time timestamp with time zone
);


ALTER TABLE public.xc_knex_migrations OWNER TO postgres;

--
-- Name: xc_knex_migrations_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.xc_knex_migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.xc_knex_migrations_id_seq OWNER TO postgres;

--
-- Name: xc_knex_migrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.xc_knex_migrations_id_seq OWNED BY public.xc_knex_migrations.id;


--
-- Name: xc_knex_migrations_lock; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.xc_knex_migrations_lock (
    index integer NOT NULL,
    is_locked integer
);


ALTER TABLE public.xc_knex_migrations_lock OWNER TO postgres;

--
-- Name: xc_knex_migrations_lock_index_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.xc_knex_migrations_lock_index_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.xc_knex_migrations_lock_index_seq OWNER TO postgres;

--
-- Name: xc_knex_migrations_lock_index_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.xc_knex_migrations_lock_index_seq OWNED BY public.xc_knex_migrations_lock.index;


--
-- Name: xc_knex_migrationsv2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.xc_knex_migrationsv2 (
    id integer NOT NULL,
    name character varying(255),
    batch integer,
    migration_time timestamp with time zone
);


ALTER TABLE public.xc_knex_migrationsv2 OWNER TO postgres;

--
-- Name: xc_knex_migrationsv2_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.xc_knex_migrationsv2_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.xc_knex_migrationsv2_id_seq OWNER TO postgres;

--
-- Name: xc_knex_migrationsv2_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.xc_knex_migrationsv2_id_seq OWNED BY public.xc_knex_migrationsv2.id;


--
-- Name: xc_knex_migrationsv2_lock; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.xc_knex_migrationsv2_lock (
    index integer NOT NULL,
    is_locked integer
);


ALTER TABLE public.xc_knex_migrationsv2_lock OWNER TO postgres;

--
-- Name: xc_knex_migrationsv2_lock_index_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.xc_knex_migrationsv2_lock_index_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.xc_knex_migrationsv2_lock_index_seq OWNER TO postgres;

--
-- Name: xc_knex_migrationsv2_lock_index_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.xc_knex_migrationsv2_lock_index_seq OWNED BY public.xc_knex_migrationsv2_lock.index;


--
-- Name: nc_acl id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_acl ALTER COLUMN id SET DEFAULT nextval('public.nc_acl_id_seq'::regclass);


--
-- Name: nc_api_tokens id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_api_tokens ALTER COLUMN id SET DEFAULT nextval('public.nc_api_tokens_id_seq'::regclass);


--
-- Name: nc_audit id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_audit ALTER COLUMN id SET DEFAULT nextval('public.nc_audit_id_seq'::regclass);


--
-- Name: nc_cron id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_cron ALTER COLUMN id SET DEFAULT nextval('public.nc_cron_id_seq'::regclass);


--
-- Name: nc_disabled_models_for_role id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_disabled_models_for_role ALTER COLUMN id SET DEFAULT nextval('public.nc_disabled_models_for_role_id_seq'::regclass);


--
-- Name: nc_evolutions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_evolutions ALTER COLUMN id SET DEFAULT nextval('public.nc_evolutions_id_seq'::regclass);


--
-- Name: nc_hooks id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_hooks ALTER COLUMN id SET DEFAULT nextval('public.nc_hooks_id_seq'::regclass);


--
-- Name: nc_loaders id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_loaders ALTER COLUMN id SET DEFAULT nextval('public.nc_loaders_id_seq'::regclass);


--
-- Name: nc_migrations id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_migrations ALTER COLUMN id SET DEFAULT nextval('public.nc_migrations_id_seq'::regclass);


--
-- Name: nc_models id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_models ALTER COLUMN id SET DEFAULT nextval('public.nc_models_id_seq'::regclass);


--
-- Name: nc_plugins id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_plugins ALTER COLUMN id SET DEFAULT nextval('public.nc_plugins_id_seq'::regclass);


--
-- Name: nc_relations id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_relations ALTER COLUMN id SET DEFAULT nextval('public.nc_relations_id_seq'::regclass);


--
-- Name: nc_resolvers id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_resolvers ALTER COLUMN id SET DEFAULT nextval('public.nc_resolvers_id_seq'::regclass);


--
-- Name: nc_roles id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_roles ALTER COLUMN id SET DEFAULT nextval('public.nc_roles_id_seq'::regclass);


--
-- Name: nc_routes id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_routes ALTER COLUMN id SET DEFAULT nextval('public.nc_routes_id_seq'::regclass);


--
-- Name: nc_rpc id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_rpc ALTER COLUMN id SET DEFAULT nextval('public.nc_rpc_id_seq'::regclass);


--
-- Name: nc_shared_bases id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_shared_bases ALTER COLUMN id SET DEFAULT nextval('public.nc_shared_bases_id_seq'::regclass);


--
-- Name: nc_shared_views id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_shared_views ALTER COLUMN id SET DEFAULT nextval('public.nc_shared_views_id_seq'::regclass);


--
-- Name: nc_store id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_store ALTER COLUMN id SET DEFAULT nextval('public.nc_store_id_seq'::regclass);


--
-- Name: xc_knex_migrations id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xc_knex_migrations ALTER COLUMN id SET DEFAULT nextval('public.xc_knex_migrations_id_seq'::regclass);


--
-- Name: xc_knex_migrations_lock index; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xc_knex_migrations_lock ALTER COLUMN index SET DEFAULT nextval('public.xc_knex_migrations_lock_index_seq'::regclass);


--
-- Name: xc_knex_migrationsv2 id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xc_knex_migrationsv2 ALTER COLUMN id SET DEFAULT nextval('public.xc_knex_migrationsv2_id_seq'::regclass);


--
-- Name: xc_knex_migrationsv2_lock index; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xc_knex_migrationsv2_lock ALTER COLUMN index SET DEFAULT nextval('public.xc_knex_migrationsv2_lock_index_seq'::regclass);


--
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.categories (category_id, category_name, description, picture) FROM stdin;
1	Beverages	Soft drinks, coffees, teas, beers, and ales	\\x
2	Condiments	Sweet and savory sauces, relishes, spreads, and seasonings	\\x
3	Confections	Desserts, candies, and sweet breads	\\x
4	Dairy Products	Cheeses	\\x
5	Grains/Cereals	Breads, crackers, pasta, and cereal	\\x
6	Meat/Poultry	Prepared meats	\\x
7	Produce	Dried fruit and bean curd	\\x
8	Seafood	Seaweed and fish	\\x
\.


--
-- Data for Name: customer_customer_demo; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.customer_customer_demo (customer_id, customer_type_id) FROM stdin;
\.


--
-- Data for Name: customer_demographics; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.customer_demographics (customer_type_id, customer_desc) FROM stdin;
\.


--
-- Data for Name: customers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.customers (customer_id, company_name, contact_name, contact_title, address, city, region, postal_code, country, phone, fax) FROM stdin;
ALFKI	Alfreds Futterkiste	Maria Anders	Sales Representative	Obere Str. 57	Berlin	\N	12209	Germany	030-0074321	030-0076545
ANATR	Ana Trujillo Emparedados y helados	Ana Trujillo	Owner	Avda. de la Constitución 2222	México D.F.	\N	05021	Mexico	(5) 555-4729	(5) 555-3745
ANTON	Antonio Moreno Taquería	Antonio Moreno	Owner	Mataderos  2312	México D.F.	\N	05023	Mexico	(5) 555-3932	\N
AROUT	Around the Horn	Thomas Hardy	Sales Representative	120 Hanover Sq.	London	\N	WA1 1DP	UK	(171) 555-7788	(171) 555-6750
BERGS	Berglunds snabbköp	Christina Berglund	Order Administrator	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden	0921-12 34 65	0921-12 34 67
BLAUS	Blauer See Delikatessen	Hanna Moos	Sales Representative	Forsterstr. 57	Mannheim	\N	68306	Germany	0621-08460	0621-08924
BLONP	Blondesddsl père et fils	Frédérique Citeaux	Marketing Manager	24, place Kléber	Strasbourg	\N	67000	France	88.60.15.31	88.60.15.32
BOLID	Bólido Comidas preparadas	Martín Sommer	Owner	C/ Araquil, 67	Madrid	\N	28023	Spain	(91) 555 22 82	(91) 555 91 99
BONAP	Bon app'	Laurence Lebihan	Owner	12, rue des Bouchers	Marseille	\N	13008	France	91.24.45.40	91.24.45.41
BOTTM	Bottom-Dollar Markets	Elizabeth Lincoln	Accounting Manager	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada	(604) 555-4729	(604) 555-3745
BSBEV	B's Beverages	Victoria Ashworth	Sales Representative	Fauntleroy Circus	London	\N	EC2 5NT	UK	(171) 555-1212	\N
CACTU	Cactus Comidas para llevar	Patricio Simpson	Sales Agent	Cerrito 333	Buenos Aires	\N	1010	Argentina	(1) 135-5555	(1) 135-4892
CENTC	Centro comercial Moctezuma	Francisco Chang	Marketing Manager	Sierras de Granada 9993	México D.F.	\N	05022	Mexico	(5) 555-3392	(5) 555-7293
CHOPS	Chop-suey Chinese	Yang Wang	Owner	Hauptstr. 29	Bern	\N	3012	Switzerland	0452-076545	\N
COMMI	Comércio Mineiro	Pedro Afonso	Sales Associate	Av. dos Lusíadas, 23	Sao Paulo	SP	05432-043	Brazil	(11) 555-7647	\N
CONSH	Consolidated Holdings	Elizabeth Brown	Sales Representative	Berkeley Gardens 12  Brewery	London	\N	WX1 6LT	UK	(171) 555-2282	(171) 555-9199
DRACD	Drachenblut Delikatessen	Sven Ottlieb	Order Administrator	Walserweg 21	Aachen	\N	52066	Germany	0241-039123	0241-059428
DUMON	Du monde entier	Janine Labrune	Owner	67, rue des Cinquante Otages	Nantes	\N	44000	France	40.67.88.88	40.67.89.89
EASTC	Eastern Connection	Ann Devon	Sales Agent	35 King George	London	\N	WX3 6FW	UK	(171) 555-0297	(171) 555-3373
ERNSH	Ernst Handel	Roland Mendel	Sales Manager	Kirchgasse 6	Graz	\N	8010	Austria	7675-3425	7675-3426
FAMIA	Familia Arquibaldo	Aria Cruz	Marketing Assistant	Rua Orós, 92	Sao Paulo	SP	05442-030	Brazil	(11) 555-9857	\N
FISSA	FISSA Fabrica Inter. Salchichas S.A.	Diego Roel	Accounting Manager	C/ Moralzarzal, 86	Madrid	\N	28034	Spain	(91) 555 94 44	(91) 555 55 93
FOLIG	Folies gourmandes	postgrese Rancé	Assistant Sales Agent	184, chaussée de Tournai	Lille	\N	59000	France	20.16.10.16	20.16.10.17
FOLKO	Folk och fä HB	Maria Larsson	Owner	Åkergatan 24	Bräcke	\N	S-844 67	Sweden	0695-34 67 21	\N
FRANK	Frankenversand	Peter Franken	Marketing Manager	Berliner Platz 43	München	\N	80805	Germany	089-0877310	089-0877451
FRANR	France restauration	Carine Schmitt	Marketing Manager	54, rue Royale	Nantes	\N	44000	France	40.32.21.21	40.32.21.20
FRANS	Franchi S.p.A.	Paolo Accorti	Sales Representative	Via Monte Bianco 34	Torino	\N	10100	Italy	011-4988260	011-4988261
FURIB	Furia Bacalhau e Frutos do Mar	Lino Rodriguez	Sales Manager	Jardim das rosas n. 32	Lisboa	\N	1675	Portugal	(1) 354-2534	(1) 354-2535
GALED	Galería del gastrónomo	Eduardo Saavedra	Marketing Manager	Rambla de Cataluña, 23	Barcelona	\N	08022	Spain	(93) 203 4560	(93) 203 4561
GODOS	Godos Cocina Típica	José Pedro Freyre	Sales Manager	C/ Romero, 33	Sevilla	\N	41101	Spain	(95) 555 82 82	\N
GOURL	Gourmet Lanchonetes	André Fonseca	Sales Associate	Av. Brasil, 442	Campinas	SP	04876-786	Brazil	(11) 555-9482	\N
GREAL	Great Lakes Food Market	Howard Snyder	Marketing Manager	2732 Baker Blvd.	Eugene	OR	97403	USA	(503) 555-7555	\N
GROSR	GROSELLA-Restaurante	Manuel Pereira	Owner	5ª Ave. Los Palos Grandes	Caracas	DF	1081	Venezuela	(2) 283-2951	(2) 283-3397
HANAR	Hanari Carnes	Mario Pontes	Accounting Manager	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil	(21) 555-0091	(21) 555-8765
HILAA	HILARION-Abastos	Carlos Hernández	Sales Representative	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela	(5) 555-1340	(5) 555-1948
HUNGC	Hungry Coyote Import Store	Yoshi Latimer	Sales Representative	City Center Plaza 516 Main St.	Elgin	OR	97827	USA	(503) 555-6874	(503) 555-2376
HUNGO	Hungry Owl All-Night Grocers	Patricia McKenna	Sales Associate	8 Johnstown Road	Cork	Co. Cork	\N	Ireland	2967 542	2967 3333
ISLAT	Island Trading	Helen Bennett	Marketing Manager	Garden House Crowther Way	Cowes	Isle of Wight	PO31 7PJ	UK	(198) 555-8888	\N
KOENE	Königlich Essen	Philip Cramer	Sales Associate	Maubelstr. 90	Brandenburg	\N	14776	Germany	0555-09876	\N
LACOR	La corne d'abondance	Daniel Tonini	Sales Representative	67, avenue de l'Europe	Versailles	\N	78000	France	30.59.84.10	30.59.85.11
LAMAI	La maison d'Asie	Annette Roulet	Sales Manager	1 rue Alsace-Lorraine	Toulouse	\N	31000	France	61.77.61.10	61.77.61.11
LAUGB	Laughing Bacchus Wine Cellars	Yoshi Tannamuri	Marketing Assistant	1900 Oak St.	Vancouver	BC	V3F 2K1	Canada	(604) 555-3392	(604) 555-7293
LAZYK	Lazy K Kountry Store	John Steel	Marketing Manager	12 Orchestra Terrace	Walla Walla	WA	99362	USA	(509) 555-7969	(509) 555-6221
LEHMS	Lehmanns Marktstand	Renate Messner	Sales Representative	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany	069-0245984	069-0245874
LETSS	Let's Stop N Shop	Jaime Yorres	Owner	87 Polk St. Suite 5	San Francisco	CA	94117	USA	(415) 555-5938	\N
LILAS	LILA-Supermercado	Carlos González	Accounting Manager	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela	(9) 331-6954	(9) 331-7256
LINOD	LINO-Delicateses	Felipe Izquierdo	Owner	Ave. 5 de Mayo Porlamar	I. de Margarita	Nueva Esparta	4980	Venezuela	(8) 34-56-12	(8) 34-93-93
LONEP	Lonesome Pine Restaurant	Fran Wilson	Sales Manager	89 Chiaroscuro Rd.	Portland	OR	97219	USA	(503) 555-9573	(503) 555-9646
MAGAA	Magazzini Alimentari Riuniti	Giovanni Rovelli	Marketing Manager	Via Ludovico il Moro 22	Bergamo	\N	24100	Italy	035-640230	035-640231
MAISD	Maison Dewey	Catherine Dewey	Sales Agent	Rue Joseph-Bens 532	Bruxelles	\N	B-1180	Belgium	(02) 201 24 67	(02) 201 24 68
MEREP	Mère Paillarde	Jean Fresnière	Marketing Assistant	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada	(514) 555-8054	(514) 555-8055
MORGK	Morgenstern Gesundkost	Alexander Feuer	Marketing Assistant	Heerstr. 22	Leipzig	\N	04179	Germany	0342-023176	\N
NORTS	North/South	Simon Crowther	Sales Associate	South House 300 Queensbridge	London	\N	SW7 1RZ	UK	(171) 555-7733	(171) 555-2530
OCEAN	Océano Atlántico Ltda.	Yvonne Moncada	Sales Agent	Ing. Gustavo Moncada 8585 Piso 20-A	Buenos Aires	\N	1010	Argentina	(1) 135-5333	(1) 135-5535
OLDWO	Old World Delicatessen	Rene Phillips	Sales Representative	2743 Bering St.	Anchorage	AK	99508	USA	(907) 555-7584	(907) 555-2880
OTTIK	Ottilies Käseladen	Henriette Pfalzheim	Owner	Mehrheimerstr. 369	Köln	\N	50739	Germany	0221-0644327	0221-0765721
PARIS	Paris spécialités	Marie Bertrand	Owner	265, boulevard Charonne	Paris	\N	75012	France	(1) 42.34.22.66	(1) 42.34.22.77
PERIC	Pericles Comidas clásicas	Guillermo Fernández	Sales Representative	Calle Dr. Jorge Cash 321	México D.F.	\N	05033	Mexico	(5) 552-3745	(5) 545-3745
PICCO	Piccolo und mehr	Georg Pipps	Sales Manager	Geislweg 14	Salzburg	\N	5020	Austria	6562-9722	6562-9723
PRINI	Princesa Isabel Vinhos	Isabel de Castro	Sales Representative	Estrada da saúde n. 58	Lisboa	\N	1756	Portugal	(1) 356-5634	\N
QUEDE	Que Delícia	Bernardo Batista	Accounting Manager	Rua da Panificadora, 12	Rio de Janeiro	RJ	02389-673	Brazil	(21) 555-4252	(21) 555-4545
QUEEN	Queen Cozinha	Lúcia Carvalho	Marketing Assistant	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil	(11) 555-1189	\N
QUICK	QUICK-Stop	Horst Kloss	Accounting Manager	Taucherstraße 10	Cunewalde	\N	01307	Germany	0372-035188	\N
RANCH	Rancho grande	Sergio Gutiérrez	Sales Representative	Av. del Libertador 900	Buenos Aires	\N	1010	Argentina	(1) 123-5555	(1) 123-5556
RATTC	Rattlesnake Canyon Grocery	Paula Wilson	Assistant Sales Representative	2817 Milton Dr.	Albuquerque	NM	87110	USA	(505) 555-5939	(505) 555-3620
REGGC	Reggiani Caseifici	Maurizio Moroni	Sales Associate	Strada Provinciale 124	Reggio Emilia	\N	42100	Italy	0522-556721	0522-556722
RICAR	Ricardo Adocicados	Janete Limeira	Assistant Sales Agent	Av. Copacabana, 267	Rio de Janeiro	RJ	02389-890	Brazil	(21) 555-3412	\N
RICSU	Richter Supermarkt	Michael Holz	Sales Manager	Grenzacherweg 237	Genève	\N	1203	Switzerland	0897-034214	\N
ROMEY	Romero y tomillo	Alejandra Camino	Accounting Manager	Gran Vía, 1	Madrid	\N	28001	Spain	(91) 745 6200	(91) 745 6210
SANTG	Santé Gourmet	Jonas Bergulfsen	Owner	Erling Skakkes gate 78	Stavern	\N	4110	Norway	07-98 92 35	07-98 92 47
SAVEA	Save-a-lot Markets	Jose Pavarotti	Sales Representative	187 Suffolk Ln.	Boise	ID	83720	USA	(208) 555-8097	\N
SEVES	Seven Seas Imports	Hari Kumar	Sales Manager	90 Wadhurst Rd.	London	\N	OX15 4NB	UK	(171) 555-1717	(171) 555-5646
SIMOB	Simons bistro	Jytte Petersen	Owner	Vinbæltet 34	Kobenhavn	\N	1734	Denmark	31 12 34 56	31 13 35 57
SPECD	Spécialités du monde	Dominique Perrier	Marketing Manager	25, rue Lauriston	Paris	\N	75016	France	(1) 47.55.60.10	(1) 47.55.60.20
SPLIR	Split Rail Beer & Ale	Art Braunschweiger	Sales Manager	P.O. Box 555	Lander	WY	82520	USA	(307) 555-4680	(307) 555-6525
SUPRD	Suprêmes délices	Pascale Cartrain	Accounting Manager	Boulevard Tirou, 255	Charleroi	\N	B-6000	Belgium	(071) 23 67 22 20	(071) 23 67 22 21
THEBI	The Big Cheese	Liz Nixon	Marketing Manager	89 Jefferson Way Suite 2	Portland	OR	97201	USA	(503) 555-3612	\N
THECR	The Cracker Box	Liu Wong	Marketing Assistant	55 Grizzly Peak Rd.	Butte	MT	59801	USA	(406) 555-5834	(406) 555-8083
TOMSP	Toms Spezialitäten	Karin Josephs	Marketing Manager	Luisenstr. 48	Münster	\N	44087	Germany	0251-031259	0251-035695
TORTU	Tortuga Restaurante	Miguel Angel Paolino	Owner	Avda. Azteca 123	México D.F.	\N	05033	Mexico	(5) 555-2933	\N
TRADH	Tradição Hipermercados	Anabela Domingues	Sales Representative	Av. Inês de Castro, 414	Sao Paulo	SP	05634-030	Brazil	(11) 555-2167	(11) 555-2168
TRAIH	Trail's Head Gourmet Provisioners	Helvetius Nagy	Sales Associate	722 DaVinci Blvd.	Kirkland	WA	98034	USA	(206) 555-8257	(206) 555-2174
VAFFE	Vaffeljernet	Palle Ibsen	Sales Manager	Smagsloget 45	Århus	\N	8200	Denmark	86 21 32 43	86 22 33 44
VICTE	Victuailles en stock	Mary Saveley	Sales Agent	2, rue du Commerce	Lyon	\N	69004	France	78.32.54.86	78.32.54.87
VINET	Vins et alcools Chevalier	Paul Henriot	Accounting Manager	59 rue de l'Abbaye	Reims	\N	51100	France	26.47.15.10	26.47.15.11
WANDK	Die Wandernde Kuh	Rita Müller	Sales Representative	Adenauerallee 900	Stuttgart	\N	70563	Germany	0711-020361	0711-035428
WARTH	Wartian Herkku	Pirkko Koskitalo	Accounting Manager	Torikatu 38	Oulu	\N	90110	Finland	981-443655	981-443655
WELLI	Wellington Importadora	Paula Parente	Sales Manager	Rua do Mercado, 12	Resende	SP	08737-363	Brazil	(14) 555-8122	\N
WHITC	White Clover Markets	Karl Jablonski	Owner	305 - 14th Ave. S. Suite 3B	Seattle	WA	98128	USA	(206) 555-4112	(206) 555-4115
WILMK	Wilman Kala	Matti Karttunen	Owner/Marketing Assistant	Keskuskatu 45	Helsinki	\N	21240	Finland	90-224 8858	90-224 8858
WOLZA	Wolski  Zajazd	Zbyszek Piestrzeniewicz	Owner	ul. Filtrowa 68	Warszawa	\N	01-012	Poland	(26) 642-7012	(26) 642-7012
\.


--
-- Data for Name: employee_territories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.employee_territories (employee_id, territory_id) FROM stdin;
1	06897
1	19713
2	01581
2	01730
2	01833
2	02116
2	02139
2	02184
2	40222
3	30346
3	31406
3	32859
3	33607
4	20852
4	27403
4	27511
5	02903
5	07960
5	08837
5	10019
5	10038
5	11747
5	14450
6	85014
6	85251
6	98004
6	98052
6	98104
7	60179
7	60601
7	80202
7	80909
7	90405
7	94025
7	94105
7	95008
7	95054
7	95060
8	19428
8	44122
8	45839
8	53404
9	03049
9	03801
9	48075
9	48084
9	48304
9	55113
9	55439
\.


--
-- Data for Name: employees; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.employees (employee_id, last_name, first_name, title, title_of_courtesy, birth_date, hire_date, address, city, region, postal_code, country, home_phone, extension, photo, notes, reports_to, photo_path) FROM stdin;
1	Davolio	Nancy	Sales Representative	Ms.	1948-12-08	1992-05-01	507 - 20th Ave. E.\\nApt. 2A	Seattle	WA	98122	USA	(206) 555-9857	5467	\\x	Education includes a BA in psychology from Colorado State University in 1970.  She also completed The Art of the Cold Call.  Nancy is a member of Toastmasters International.	2	http://accweb/emmployees/davolio.bmp
2	Fuller	Andrew	Vice President, Sales	Dr.	1952-02-19	1992-08-14	908 W. Capital Way	Tacoma	WA	98401	USA	(206) 555-9482	3457	\\x	Andrew received his BTS commercial in 1974 and a Ph.D. in international marketing from the University of Dallas in 1981.  He is fluent in French and Italian and reads German.  He joined the company as a sales representative, was promoted to sales manager in January 1992 and to vice president of sales in March 1993.  Andrew is a member of the Sales Management Roundtable, the Seattle Chamber of Commerce, and the Pacific Rim Importers Association.	\N	http://accweb/emmployees/fuller.bmp
3	Leverling	Janet	Sales Representative	Ms.	1963-08-30	1992-04-01	722 Moss Bay Blvd.	Kirkland	WA	98033	USA	(206) 555-3412	3355	\\x	Janet has a BS degree in chemistry from Boston College (1984).  She has also completed a certificate program in food retailing management.  Janet was hired as a sales associate in 1991 and promoted to sales representative in February 1992.	2	http://accweb/emmployees/leverling.bmp
4	Peacock	Margaret	Sales Representative	Mrs.	1937-09-19	1993-05-03	4110 Old Redmond Rd.	Redmond	WA	98052	USA	(206) 555-8122	5176	\\x	Margaret holds a BA in English literature from Concordia College (1958) and an MA from the American Institute of Culinary Arts (1966).  She was assigned to the London office temporarily from July through November 1992.	2	http://accweb/emmployees/peacock.bmp
5	Buchanan	Steven	Sales Manager	Mr.	1955-03-04	1993-10-17	14 Garrett Hill	London	\N	SW1 8JR	UK	(71) 555-4848	3453	\\x	Steven Buchanan graduated from St. Andrews University, Scotland, with a BSC degree in 1976.  Upon joining the company as a sales representative in 1992, he spent 6 months in an orientation program at the Seattle office and then returned to his permanent post in London.  He was promoted to sales manager in March 1993.  Mr. Buchanan has completed the courses Successful Telemarketing and International Sales Management.  He is fluent in French.	2	http://accweb/emmployees/buchanan.bmp
6	Suyama	Michael	Sales Representative	Mr.	1963-07-02	1993-10-17	Coventry House\\nMiner Rd.	London	\N	EC2 7JR	UK	(71) 555-7773	428	\\x	Michael is a graduate of Sussex University (MA, economics, 1983) and the University of California at Los Angeles (MBA, marketing, 1986).  He has also taken the courses Multi-Cultural Selling and Time Management for the Sales Professional.  He is fluent in Japanese and can read and write French, Portuguese, and Spanish.	5	http://accweb/emmployees/davolio.bmp
7	King	Robert	Sales Representative	Mr.	1960-05-29	1994-01-02	Edgeham Hollow\\nWinchester Way	London	\N	RG1 9SP	UK	(71) 555-5598	465	\\x	Robert King served in the Peace Corps and traveled extensively before completing his degree in English at the University of Michigan in 1992, the year he joined the company.  After completing a course entitled Selling in Europe, he was transferred to the London office in March 1993.	5	http://accweb/emmployees/davolio.bmp
8	Callahan	Laura	Inside Sales Coordinator	Ms.	1958-01-09	1994-03-05	4726 - 11th Ave. N.E.	Seattle	WA	98105	USA	(206) 555-1189	2344	\\x	Laura received a BA in psychology from the University of Washington.  She has also completed a course in business French.  She reads and writes French.	2	http://accweb/emmployees/davolio.bmp
9	Dodsworth	Anne	Sales Representative	Ms.	1966-01-27	1994-11-15	7 Houndstooth Rd.	London	\N	WG2 7LT	UK	(71) 555-4444	452	\\x	Anne has a BA degree in English from St. Lawrence College.  She is fluent in French and German.	5	http://accweb/emmployees/davolio.bmp
\.


--
-- Data for Name: nc_acl; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_acl (id, project_id, db_alias, tn, acl, type, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_api_tokens; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_api_tokens (id, project_id, db_alias, description, permissions, token, expiry, enabled, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_audit; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_audit (id, "user", ip, project_id, db_alias, model_name, model_id, op_type, op_sub_type, status, description, details, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_audit_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_audit_v2 (id, "user", ip, base_id, project_id, fk_model_id, row_id, op_type, op_sub_type, status, description, details, created_at, updated_at) FROM stdin;
adt_sqzwc3j9ao2rm6	postgreserikjonsson@yahoo.fr	::1	\N	\N	\N	\N	AUTHENTICATION	SIGNUP	\N	signed up 	\N	2022-05-18 11:23:15.918236+02	2022-05-18 11:23:15.918236+02
adt_61af8sdq6xayoq	postgreserikjonsson@yahoo.fr	::1	\N	p_41297240e6of48	md_r7cgbvzh9qy7qk	10250	DATA	UPDATE	\N	Table orders : field ShipName got changed from  Hanari Carnes to Hanari Carnes	<span class="">ShipName</span>\n  : <span class="text-decoration-line-through red px-2 lighten-4 black--text">Hanari Carnes</span>\n  <span class="black--text green lighten-4 px-2">Hanari Carnes</span>	2022-05-18 11:31:46.570156+02	2022-05-18 11:31:46.570156+02
adt_4o23i9mh3un7ou	postgreserikjonsson@yahoo.fr	::1	\N	p_41297240e6of48	\N	\N	TABLE_COLUMN	CREATED	\N	created column undefined with alias title19 from table orders	\N	2022-05-18 11:58:47.555407+02	2022-05-18 11:58:47.555407+02
adt_9a1xtge5ygm9ay	postgreserikjonsson@yahoo.fr	::1	\N	p_41297240e6of48	\N	\N	TABLE_COLUMN	UPDATED	\N	updated column null with alias title19 from table orders	\N	2022-05-18 11:59:05.78394+02	2022-05-18 11:59:05.78394+02
adt_d7mw79rw5qbm8l	postgreserikjonsson@yahoo.fr	::1	\N	p_41297240e6of48	\N	\N	TABLE_COLUMN	UPDATED	\N	updated column null with alias title19 from table orders	\N	2022-05-18 12:00:24.246653+02	2022-05-18 12:00:24.246653+02
adt_12q596fcruileh	postgreserikjonsson@yahoo.fr	::1	\N	p_41297240e6of48	\N	\N	TABLE_COLUMN	UPDATED	\N	updated column null with alias test from table orders	\N	2022-05-18 12:01:05.453586+02	2022-05-18 12:01:05.453586+02
\.


--
-- Data for Name: nc_bases_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_bases_v2 (id, project_id, alias, config, meta, is_meta, type, inflection_column, inflection_table, created_at, updated_at) FROM stdin;
ds_aej9ao4mzw9cff	p_41297240e6of48	\N	U2FsdGVkX190YFEYo2QOytteAKHeovDdd7+okN/otw7w1MyKlJYjRNIN0Jxoanh/3rhlQILn3mvfpVzo9ef2YgJ/GZsVzMrGDkrw90zvch+SDLFTJIqrat0Xt1fd6WTrgk/ah4E8VW1pbPV34/thvsnyGK7QRWyrI8J2/pcyF5ftF6ih7MaqH95ZwFRjaszLNxCbo0Smb+XFZheqgBjfI5uJYLkECloXF7hV24pSJu8VknARd25hkoV/9f+mZoSVfb5wNdiwoCzYVrWRk5sbMtiPepsZUwzr7hr5FR2qoTGoWqPVJigG1sDAES5YdfPXSoVS6Ofd56d04OwDXT2mCV/L0cgWum392H+PW9a59xokvDjl6nrNzbZgqh+cuNkviqnwqaS/J00Ib2YMRbnCffzUehUSNYJefvN4wuP4RbRBC3Re9U/v2Y7wXyNziq7s	\N	\N	pg	camelize	camelize	2022-05-18 11:24:10.786224+02	2022-05-18 11:24:10.786224+02
\.


--
-- Data for Name: nc_col_formula_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_col_formula_v2 (id, fk_column_id, formula, formula_raw, error, deleted, "order", created_at, updated_at) FROM stdin;
fm_pihkbeir0674nf	cl_pjm18dcg199nid	DATEADD(cl_42mxp318j7y76i, 2, "day")	DATEADD(OrderDate, 2, "day")	\N	\N	\N	2022-05-18 12:01:05.451748+02	2022-05-18 12:01:05.451748+02
\.


--
-- Data for Name: nc_col_lookup_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_col_lookup_v2 (id, fk_column_id, fk_relation_column_id, fk_lookup_column_id, deleted, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_col_relations_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_col_relations_v2 (id, ref_db_alias, type, virtual, db_type, fk_column_id, fk_related_model_id, fk_child_column_id, fk_parent_column_id, fk_mm_model_id, fk_mm_child_column_id, fk_mm_parent_column_id, ur, dr, fk_index_name, deleted, created_at, updated_at) FROM stdin;
ln_i3b1jj5gz1omsv	\N	bt	\N	\N	cl_tbvbq12n5xrtso	md_71j5r6adaq9l3i	cl_7vy94bkvan3b13	cl_9wjh8ccwe2bpmz	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.901808+02	2022-05-18 11:24:12.901808+02
ln_dka88ujxjn0hxk	\N	hm	\N	\N	cl_9w6yddj9cco5ke	md_7vpsxklk3sph75	cl_7vy94bkvan3b13	cl_9wjh8ccwe2bpmz	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.901974+02	2022-05-18 11:24:12.901974+02
ln_crf3dcv2nl4quu	\N	bt	\N	\N	cl_cz3l6lgwksaq2e	md_yor91fufon7h6c	cl_k5ansjw5dhsxmw	cl_99h5zfuj8v82zf	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.902215+02	2022-05-18 11:24:12.902215+02
ln_5s4tn1g02tui7o	\N	hm	\N	\N	cl_dyz994lxx7yesp	md_r7cgbvzh9qy7qk	cl_lxtgajo4s3br49	cl_u7wce2nao3rabj	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.902733+02	2022-05-18 11:24:12.902733+02
ln_qq0fi4f11ee0sa	\N	hm	\N	\N	cl_t3dlbbwzew7sam	md_q2z56jvps22ukx	cl_0qe00wlxhp6l8d	cl_6i16gsjxn2t5bh	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.90292+02	2022-05-18 11:24:12.90292+02
ln_zyh8amvm4k7sxq	\N	hm	\N	\N	cl_a7kx6brjog1as9	md_mmp2o2atqtudsn	cl_rk5bzouixq4oyt	cl_gpkukq2p7xpr9n	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.903136+02	2022-05-18 11:24:12.903136+02
ln_mkyeugxaju9mkf	\N	hm	\N	\N	cl_4gstsln64mwssn	md_89v98o1rpi4idv	cl_mhh2pb9uvhqu8d	cl_pvlhgpc2mbanhq	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.903868+02	2022-05-18 11:24:12.903868+02
ln_rirck56jum3q9e	\N	bt	\N	\N	cl_nqzdipc2k8fwu9	md_r7cgbvzh9qy7qk	cl_z4jql0g2ge0i9x	cl_lwxz8ubw0dlz2l	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.904115+02	2022-05-18 11:24:12.904115+02
ln_lbsk8o3lgd3wa6	\N	hm	\N	\N	cl_318j4uypmgfmdn	md_thf0xauk4e6qxs	cl_r39qpsifh4dd16	cl_c0p5yzzp5h2fdv	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.904376+02	2022-05-18 11:24:12.904376+02
ln_6wj20j2859hoqk	\N	bt	\N	\N	cl_d8fg3jivyrrhux	md_ue8cshscw72tz4	cl_narv6h4qk6asbn	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.904731+02	2022-05-18 11:24:12.904731+02
ln_wyqw2nz04d12sw	\N	hm	\N	\N	cl_dic1zbej0ei7x4	md_j409llgiqo90yz	cl_fifa1pn2w5hhiv	cl_vvtlvtc6gzirw0	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.905087+02	2022-05-18 11:24:12.905087+02
ln_hz050mdzwsmlg7	\N	bt	\N	\N	cl_art81q63geekql	md_ue8cshscw72tz4	cl_dpxgtdswb10tkh	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.905276+02	2022-05-18 11:24:12.905276+02
ln_4cvl0mbamosb2l	\N	bt	\N	\N	cl_mr8rz9hd7ttjwp	md_dg3rmvfk4bsg1i	cl_jprcavrmue3v7e	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.905519+02	2022-05-18 11:24:12.905519+02
ln_lyxk4rua6ywvw8	\N	bt	\N	\N	cl_ltn26rjondvorz	md_ue8cshscw72tz4	cl_ycnsnkvz4v8akf	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.906135+02	2022-05-18 11:24:12.906135+02
ln_jvln8ma0k8ikb6	\N	bt	\N	\N	cl_wrshd8cmtqyzk5	md_nj2q6tfni3b5a8	cl_mhh2pb9uvhqu8d	cl_pvlhgpc2mbanhq	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.906377+02	2022-05-18 11:24:12.906377+02
ln_wjlzp4tz8k8ysk	\N	bt	\N	\N	cl_k8gnkl4qn9ylb2	md_nj2q6tfni3b5a8	cl_h2is03dfyemdhi	cl_pvlhgpc2mbanhq	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.906664+02	2022-05-18 11:24:12.906664+02
ln_kwkmix68vx7gbs	\N	hm	\N	\N	cl_go4hvqdi86lz0l	md_7vpsxklk3sph75	cl_jta9jl2qws2s1t	cl_bq67f15udwlp8d	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.906833+02	2022-05-18 11:24:12.906833+02
ln_imw9qt8h1v0lzm	\N	hm	\N	\N	cl_jmagj4k4jx5ffw	md_hle53jhinff293	cl_nv2qb0bqfi14xj	cl_ydjxav6o2pfyt2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.907019+02	2022-05-18 11:24:12.907019+02
ln_16g8zlcmznukhz	\N	hm	\N	\N	cl_j44zcituyfu32m	md_ejwlk09y393nhf	cl_2b7kd2rxaa6i09	cl_py299lng6pd9y4	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.907205+02	2022-05-18 11:24:12.907205+02
ln_b1x6wk1xgl0i36	\N	bt	\N	\N	cl_6fou6htyzx6fa8	md_ue8cshscw72tz4	cl_cyde2oo9zyylv7	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.907468+02	2022-05-18 11:24:12.907468+02
ln_oha4bslwwjs5hx	\N	bt	\N	\N	cl_xo0892py3ndpke	md_ue8cshscw72tz4	cl_2l9fbdjfosuadr	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.907641+02	2022-05-18 11:24:12.907641+02
ln_6hx7ttsj5fkn5h	\N	bt	\N	\N	cl_2ruapfr842h3bp	md_6ht8y7vgzcek0p	cl_5l0iqbng3qv1wo	cl_1sizugbjggamtt	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.907941+02	2022-05-18 11:24:12.907941+02
ln_u76hrwyicuaw63	\N	bt	\N	\N	cl_3rcpv6o301oikz	md_dg3rmvfk4bsg1i	cl_ygmn763v7bzofp	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.908298+02	2022-05-18 11:24:12.908298+02
ln_0gj7vksn2mhbau	\N	bt	\N	\N	cl_8cezepfnbj7t2k	md_ue8cshscw72tz4	cl_pi3yol2gqnb28k	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.908545+02	2022-05-18 11:24:12.908545+02
ln_yjsyn2ekrnu4vw	\N	hm	\N	\N	cl_v9cnldmlymawdy	md_s9jny9njrhdnwz	cl_u6d2985s4k2vje	cl_68x1go22132wqm	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.908792+02	2022-05-18 11:24:12.908792+02
ln_8am35ny7x4pope	\N	bt	\N	\N	cl_d0if632wlv7hiu	md_yi49xujy865heu	cl_nv2qb0bqfi14xj	cl_ydjxav6o2pfyt2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.909123+02	2022-05-18 11:24:12.909123+02
ln_euwdxuogcx0krd	\N	hm	\N	\N	cl_u1wxcvmwmn6urv	md_gwspodfugzyzs2	cl_m5ihdmbagtnr1m	cl_yuc04l9uriu78e	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.909284+02	2022-05-18 11:24:12.909284+02
ln_oidxnkg5fh753v	\N	bt	\N	\N	cl_p619rqzumc5ueb	md_ue8cshscw72tz4	cl_6yy0dvgw2yb248	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.909505+02	2022-05-18 11:24:12.909505+02
ln_uhhmvr374ht42d	\N	bt	\N	\N	cl_7s13f8gdnr3s0p	md_ue8cshscw72tz4	cl_jg8un0z5v4mbq0	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.909788+02	2022-05-18 11:24:12.909788+02
ln_kqkdc2ht9plzwq	\N	hm	\N	\N	cl_sxpcp2cwk39cag	md_hle53jhinff293	cl_fikkm6avrnvalf	cl_eqywr4zelb2x2k	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.910235+02	2022-05-18 11:24:12.910235+02
ln_qj0t1e96nnkijl	\N	hm	\N	\N	cl_fzz2wpiidlwgie	md_91fcdvmtveuz77	cl_jprcavrmue3v7e	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.910727+02	2022-05-18 11:24:12.910727+02
ln_s0xjr6s3i8fho8	\N	hm	\N	\N	cl_j32rfmzk6jy8en	md_op8dq4us6zg5tn	cl_8vqk98wddq5tto	cl_da7co98ztmz4v6	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.911218+02	2022-05-18 11:24:12.911218+02
ln_eaxzt6m1143kwu	\N	hm	\N	\N	cl_6skr9pfy1fpbdh	md_mmp2o2atqtudsn	cl_pwh6s0lomxswl7	cl_ji20zt281pec2c	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.911759+02	2022-05-18 11:24:12.911759+02
ln_5uf7hei46vlyb2	\N	bt	\N	\N	cl_95l3l24r682gvn	md_ue8cshscw72tz4	cl_2jbc6l8uccbcxv	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.912336+02	2022-05-18 11:24:12.912336+02
ln_jms71yk4rx6xqk	\N	hm	\N	\N	cl_9q88yelvrwjb6u	md_hle53jhinff293	cl_t92sq868h7db4w	cl_hl2a11v360e4b1	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.912826+02	2022-05-18 11:24:12.912826+02
ln_kpy47iq29at1ng	\N	hm	\N	\N	cl_h44rjic2ztmxeb	md_q2z56jvps22ukx	cl_k5ansjw5dhsxmw	cl_99h5zfuj8v82zf	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.913145+02	2022-05-18 11:24:12.913145+02
ln_qktrh9i3dkd4nx	\N	bt	\N	\N	cl_mg31nh0giwezx0	md_ue8cshscw72tz4	cl_dri9nl4xksb0u6	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.913375+02	2022-05-18 11:24:12.913375+02
ln_xt2t95ikmy8bgf	\N	hm	\N	\N	cl_wzx2z9fw9kc75l	md_s9jny9njrhdnwz	cl_z4jql0g2ge0i9x	cl_lwxz8ubw0dlz2l	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.913586+02	2022-05-18 11:24:12.913586+02
ln_oo59fbqsmxw47x	\N	hm	\N	\N	cl_m2p63veyy3cmww	md_tm1aur9f3yn2l2	cl_ycnsnkvz4v8akf	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.913782+02	2022-05-18 11:24:12.913782+02
ln_4qko5d5bvn8lm5	\N	hm	\N	\N	cl_vr6w15ow265mdm	md_ebe9bibmq1l2uf	cl_5l0iqbng3qv1wo	cl_1sizugbjggamtt	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.913988+02	2022-05-18 11:24:12.913988+02
ln_k47mzfy5j0l1w1	\N	hm	\N	\N	cl_ftb3421h4x8xxr	md_whwvfz6cb3ljdb	cl_0cp6rgcqobi280	cl_v2ll0vs8jz4501	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.914206+02	2022-05-18 11:24:12.914206+02
ln_win883oep8kuxl	\N	hm	\N	\N	cl_9rx74ga8teuo7x	md_gwspodfugzyzs2	cl_rfzkjum97sm4px	cl_zdgsrgtaf7i8jh	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:12.915029+02	2022-05-18 11:24:12.915029+02
ln_wwtbeesij8iamu	\N	bt	\N	\N	cl_63p3m5hmmo1sal	md_vb4nqekkid2rl6	cl_jta9jl2qws2s1t	cl_bq67f15udwlp8d	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.004586+02	2022-05-18 11:24:13.004586+02
ln_xtkvvsx4baqkhg	\N	bt	\N	\N	cl_yuvf9ho2j6gyx2	md_j409llgiqo90yz	cl_0qe00wlxhp6l8d	cl_6i16gsjxn2t5bh	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.004792+02	2022-05-18 11:24:13.004792+02
ln_m5me4vsgyy892c	\N	hm	\N	\N	cl_nbodzf77l55cru	md_r7cgbvzh9qy7qk	cl_48210hus0kkb2v	cl_bq67f15udwlp8d	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.005755+02	2022-05-18 11:24:13.005755+02
ln_l0hy0uw1d2xex3	\N	bt	\N	\N	cl_msju6cxj8nqjm4	md_jgq6gbzvyfbk2g	cl_fifa1pn2w5hhiv	cl_vvtlvtc6gzirw0	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.006006+02	2022-05-18 11:24:13.006006+02
ln_7lir7egpkx4ab6	\N	bt	\N	\N	cl_38ziv4dkkdh8iw	md_dg3rmvfk4bsg1i	cl_py299lng6pd9y4	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.006203+02	2022-05-18 11:24:13.006203+02
ln_vbvkok1f1dgfmx	\N	bt	\N	\N	cl_8bq9golhirsif4	md_ue8cshscw72tz4	cl_5gkj8u9n2irl5t	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.006533+02	2022-05-18 11:24:13.006533+02
ln_il65cfxkbigptb	\N	hm	\N	\N	cl_5szjsumbk0yl6e	md_yi49xujy865heu	cl_bg6e57ixpkc66e	cl_eqywr4zelb2x2k	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.009059+02	2022-05-18 11:24:13.009059+02
ln_owgz7e882a4ty0	\N	hm	\N	\N	cl_aedmuthbh2dyrz	md_gwspodfugzyzs2	cl_3gm2h79uxhdzcc	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.012706+02	2022-05-18 11:24:13.012706+02
ln_x5ir6moy9ubzrk	\N	hm	\N	\N	cl_hi074lxlm7gyan	md_89v98o1rpi4idv	cl_t4dwkmgadehh7w	cl_1sizugbjggamtt	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.014846+02	2022-05-18 11:24:13.014846+02
ln_wni5n8i1oyqrnm	\N	bt	\N	\N	cl_rrv3kdmoalscjb	md_ue8cshscw72tz4	cl_cxqf2kqwmat0k3	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.073714+02	2022-05-18 11:24:13.073714+02
ln_r7omgr64g92o73	\N	hm	\N	\N	cl_18yvnok8ybldha	md_5hjnvbjhgfiz6r	cl_p0r4ptz7y81v5n	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.076549+02	2022-05-18 11:24:13.076549+02
ln_leipvkixea4ps4	\N	bt	\N	\N	cl_ru32bw7buii6la	md_yor91fufon7h6c	cl_tqy34owdpe210n	cl_99h5zfuj8v82zf	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.106675+02	2022-05-18 11:24:13.106675+02
ln_4wrwfaekwvimx6	\N	bt	\N	\N	cl_xm8dld9std1g53	md_2vqez05sn215dj	cl_0cp6rgcqobi280	cl_v2ll0vs8jz4501	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.006802+02	2022-05-18 11:24:13.006802+02
ln_n6g02q5du61uas	\N	bt	\N	\N	cl_6dhtrez796pgcx	md_ue8cshscw72tz4	cl_p0r4ptz7y81v5n	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.009461+02	2022-05-18 11:24:13.009461+02
ln_4bps53kow68db8	\N	bt	\N	\N	cl_l7p1sgvsfxg2m2	md_dg3rmvfk4bsg1i	cl_da7co98ztmz4v6	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.012892+02	2022-05-18 11:24:13.012892+02
ln_s3b4gspeuywe86	\N	bt	\N	\N	cl_5o0i7pgrh2i5oh	md_ue8cshscw72tz4	cl_rzvj3urxhsfmli	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.077259+02	2022-05-18 11:24:13.077259+02
ln_pypmv8wvd2ug6e	\N	hm	\N	\N	cl_fpxyzzfs6zlpdw	md_2vqez05sn215dj	cl_v2ll0vs8jz4501	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.106904+02	2022-05-18 11:24:13.106904+02
ln_fhtmd9dvdsnnie	\N	hm	\N	\N	cl_jtqvujcd57ffc1	md_zmhiordb5vpwkf	cl_c0p5yzzp5h2fdv	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.139357+02	2022-05-18 11:24:13.139357+02
ln_r6riyqexlbd9u8	\N	hm	\N	\N	cl_7yabknvt1yvb2p	md_ajkugjxa4haf22	cl_py299lng6pd9y4	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.154591+02	2022-05-18 11:24:13.154591+02
ln_v3edkjb2o4bigg	\N	hm	\N	\N	cl_87tqdqa6v32q7t	md_12azccujctejop	cl_ygmn763v7bzofp	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.168151+02	2022-05-18 11:24:13.168151+02
ln_0860q2yq7fw1ad	\N	hm	\N	\N	cl_vbpxe8p4sakrfa	md_wqy1a8c5ggpyea	cl_enjcwnqd9lm0hl	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.175409+02	2022-05-18 11:24:13.175409+02
ln_fadgp4hsocpkyx	\N	bt	\N	\N	cl_zc7rvd1ff56cib	md_cltahxtbesn8vs	cl_2l1uwu50cen5px	cl_hl2a11v360e4b1	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.183575+02	2022-05-18 11:24:13.183575+02
ln_649sp99j2j0kbx	\N	bt	\N	\N	cl_oe5982qxpft694	md_79w8haiknxbkzd	cl_rk5bzouixq4oyt	cl_gpkukq2p7xpr9n	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.007309+02	2022-05-18 11:24:13.007309+02
ln_gz6ohoi1lcxb8a	\N	bt	\N	\N	cl_6nefo9nk04gtus	md_ue8cshscw72tz4	cl_v63cf04xvz97pb	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.013457+02	2022-05-18 11:24:13.013457+02
ln_fy2pf0xigpngta	\N	bt	\N	\N	cl_uq57oat6mrcthd	md_mmp2o2atqtudsn	cl_u6d2985s4k2vje	cl_68x1go22132wqm	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.007526+02	2022-05-18 11:24:13.007526+02
ln_zaz0ginscf7l8i	\N	bt	\N	\N	cl_leidqf2j9vfghu	md_zmhiordb5vpwkf	cl_r39qpsifh4dd16	cl_c0p5yzzp5h2fdv	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.011127+02	2022-05-18 11:24:13.011127+02
ln_2vaxap12p76jiv	\N	hm	\N	\N	cl_lm7p4u967pdvrv	md_ruvc4yz1y8yonz	cl_fl7eap9v6vs6k5	cl_hl2a11v360e4b1	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.013758+02	2022-05-18 11:24:13.013758+02
ln_4srikf4vhnyvw6	\N	bt	\N	\N	cl_iih5ukipx3n3mn	md_ue8cshscw72tz4	cl_luguf02s068frs	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.072452+02	2022-05-18 11:24:13.072452+02
ln_lgaet5ma24ytvb	\N	bt	\N	\N	cl_2jgt0lwyo55epa	md_eqrk8yd76gurbf	cl_pwh6s0lomxswl7	cl_ji20zt281pec2c	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.074809+02	2022-05-18 11:24:13.074809+02
ln_g6uhq4duz6yipd	\N	hm	\N	\N	cl_s0w0vr1psqp3jm	md_ue8cshscw72tz4	cl_wn3aloaldty24k	cl_hl2a11v360e4b1	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.107203+02	2022-05-18 11:24:13.107203+02
ln_fqyexl84nka320	\N	hm	\N	\N	cl_maydm3rcbofz04	md_muq975426eve8k	cl_w9k4oudkzve2tu	cl_hl2a11v360e4b1	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.127143+02	2022-05-18 11:24:13.127143+02
ln_4bhowxqx7cvz0g	\N	hm	\N	\N	cl_p1fuzr7pw5w24g	md_dg3rmvfk4bsg1i	cl_2l1uwu50cen5px	cl_hl2a11v360e4b1	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.140467+02	2022-05-18 11:24:13.140467+02
ln_u94174ucenpwoh	\N	bt	\N	\N	cl_vo3yhpdpc37tle	md_yi49xujy865heu	cl_o1e6spelqpafwg	cl_ydjxav6o2pfyt2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.150779+02	2022-05-18 11:24:13.150779+02
ln_jnht096kxwsh20	\N	bt	\N	\N	cl_nv3xs219vjcf81	md_ofa8d1ocm4oowx	cl_kmwlqydz7qisje	cl_eqywr4zelb2x2k	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.160458+02	2022-05-18 11:24:13.160458+02
ln_dd67hl43rf5p5p	\N	hm	\N	\N	cl_hsqf01u6x6uwz7	md_0z5ofu8o9xph1j	cl_h2is03dfyemdhi	cl_pvlhgpc2mbanhq	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.007905+02	2022-05-18 11:24:13.007905+02
ln_srqlbgygbc5acn	\N	bt	\N	\N	cl_ikur0598w6hz3f	md_cltahxtbesn8vs	cl_t92sq868h7db4w	cl_hl2a11v360e4b1	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.011668+02	2022-05-18 11:24:13.011668+02
ln_an2vkge502vjnx	\N	bt	\N	\N	cl_03jw5hlwoiw2ji	md_yor91fufon7h6c	cl_2n98v3z6z9y6eo	cl_99h5zfuj8v82zf	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.075098+02	2022-05-18 11:24:13.075098+02
ln_etcbhikwmp6ie9	\N	bt	\N	\N	cl_i8optm3oaj6xqp	md_dg3rmvfk4bsg1i	cl_v2ll0vs8jz4501	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.07886+02	2022-05-18 11:24:13.07886+02
ln_jv3ffve924jiry	\N	hm	\N	\N	cl_j5ssutoh7jpsjp	md_5hjnvbjhgfiz6r	cl_cxqf2kqwmat0k3	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.10769+02	2022-05-18 11:24:13.10769+02
ln_9jubk0nvsd1hva	\N	hm	\N	\N	cl_wjjizc3l97gv2y	md_ruvc4yz1y8yonz	cl_dri9nl4xksb0u6	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.127959+02	2022-05-18 11:24:13.127959+02
ln_ufqp3rid4lkzha	\N	hm	\N	\N	cl_nhb4fulxib5qgn	md_ruvc4yz1y8yonz	cl_s38bav6sh7c4en	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.140657+02	2022-05-18 11:24:13.140657+02
ln_lxjoa0wxccmxh1	\N	hm	\N	\N	cl_ll6jzhxjga93yn	md_ruvc4yz1y8yonz	cl_rzvj3urxhsfmli	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.158656+02	2022-05-18 11:24:13.158656+02
ln_41urhvkf8jt0kw	\N	hm	\N	\N	cl_nht4pi4vgcjvr2	md_ruvc4yz1y8yonz	cl_87e0ickdzoj44q	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.170316+02	2022-05-18 11:24:13.170316+02
ln_rrd7o43u3xzcez	\N	hm	\N	\N	cl_s55zpov56wgfmh	md_ruvc4yz1y8yonz	cl_dp5y1a7pcz0s54	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.176213+02	2022-05-18 11:24:13.176213+02
ln_8odu26jlxx6ksy	\N	hm	\N	\N	cl_200i4b3amjhxyg	md_rmg6p4gh772ip8	cl_dpxgtdswb10tkh	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.184172+02	2022-05-18 11:24:13.184172+02
ln_lt7nb42xqzmkaw	\N	hm	\N	\N	cl_stxhadl7xlmpqv	md_rmg6p4gh772ip8	cl_5gkj8u9n2irl5t	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.18975+02	2022-05-18 11:24:13.18975+02
ln_ff7n2km1ktz6zi	\N	hm	\N	\N	cl_meczg88bpm9dk2	md_rmg6p4gh772ip8	cl_luguf02s068frs	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.196747+02	2022-05-18 11:24:13.196747+02
ln_whtqe5p3ym3yhc	\N	hm	\N	\N	cl_s0x15kjbe4c2at	md_6pjpqivc2ymh3b	cl_narv6h4qk6asbn	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.20265+02	2022-05-18 11:24:13.20265+02
ln_yzzi173ioxu8m5	\N	hm	\N	\N	cl_et3x24qyhtam8z	md_gwspodfugzyzs2	cl_v63cf04xvz97pb	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.207993+02	2022-05-18 11:24:13.207993+02
ln_u8jh8gj59q7j9h	\N	hm	\N	\N	cl_iaf4ewburqa3fd	md_op8dq4us6zg5tn	cl_6yy0dvgw2yb248	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.215382+02	2022-05-18 11:24:13.215382+02
ln_et6ppy1w6ypey6	\N	hm	\N	\N	cl_ob4dqzwkr599rk	md_whwvfz6cb3ljdb	cl_2l9fbdjfosuadr	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.22133+02	2022-05-18 11:24:13.22133+02
ln_jggqgoi96hktx7	\N	hm	\N	\N	cl_83lo23d61k9mdi	md_2vqez05sn215dj	cl_yaupebwbnv4qxt	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.228815+02	2022-05-18 11:24:13.228815+02
ln_pemtba1qyhmjf2	\N	hm	\N	\N	cl_h1zilufq67bh99	md_thf0xauk4e6qxs	cl_2jbc6l8uccbcxv	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.235257+02	2022-05-18 11:24:13.235257+02
ln_qrqa83gqh67ig2	\N	hm	\N	\N	cl_lbsn3zcyhy5knh	md_ejwlk09y393nhf	cl_jg8un0z5v4mbq0	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.241051+02	2022-05-18 11:24:13.241051+02
ln_5saadowjl6wu60	\N	hm	\N	\N	cl_d13oki8bjyop85	md_wqy1a8c5ggpyea	cl_pi3yol2gqnb28k	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.248791+02	2022-05-18 11:24:13.248791+02
ln_dzhz7bpxw5xsmu	\N	bt	\N	\N	cl_kh8758rfbidubl	md_cltahxtbesn8vs	cl_wn3aloaldty24k	cl_hl2a11v360e4b1	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.254625+02	2022-05-18 11:24:13.254625+02
ln_iqy959q24n12j9	\N	mm	\N	\N	cl_xzi1g0bfxea38l	md_vb4nqekkid2rl6	cl_9wjh8ccwe2bpmz	cl_bq67f15udwlp8d	md_7vpsxklk3sph75	cl_7vy94bkvan3b13	cl_jta9jl2qws2s1t	\N	\N	\N	\N	2022-05-18 11:24:13.261554+02	2022-05-18 11:24:13.261554+02
ln_yrenaiixfcwqvg	\N	mm	\N	\N	cl_xvgp8vm7xginj2	md_71j5r6adaq9l3i	cl_bq67f15udwlp8d	cl_9wjh8ccwe2bpmz	md_7vpsxklk3sph75	cl_jta9jl2qws2s1t	cl_7vy94bkvan3b13	\N	\N	\N	\N	2022-05-18 11:24:13.267093+02	2022-05-18 11:24:13.267093+02
ln_fr7qpjvqwvg1zn	\N	mm	\N	\N	cl_n0gu1q86ugzp2h	md_j409llgiqo90yz	cl_99h5zfuj8v82zf	cl_6i16gsjxn2t5bh	md_q2z56jvps22ukx	cl_k5ansjw5dhsxmw	cl_0qe00wlxhp6l8d	\N	\N	\N	\N	2022-05-18 11:24:13.277291+02	2022-05-18 11:24:13.277291+02
ln_hx9tzmmie9eqf0	\N	mm	\N	\N	cl_hbp7w3tjwejjua	md_yor91fufon7h6c	cl_6i16gsjxn2t5bh	cl_99h5zfuj8v82zf	md_q2z56jvps22ukx	cl_0qe00wlxhp6l8d	cl_k5ansjw5dhsxmw	\N	\N	\N	\N	2022-05-18 11:24:13.283217+02	2022-05-18 11:24:13.283217+02
ln_d82pvtmm36u4ns	\N	mm	\N	\N	cl_b7z0ydp2g1nxu9	md_6ht8y7vgzcek0p	cl_pvlhgpc2mbanhq	cl_1sizugbjggamtt	md_89v98o1rpi4idv	cl_mhh2pb9uvhqu8d	cl_t4dwkmgadehh7w	\N	\N	\N	\N	2022-05-18 11:24:13.299252+02	2022-05-18 11:24:13.299252+02
ln_utsnq5lwhmzq7f	\N	mm	\N	\N	cl_ezsdkn1opttaz3	md_nj2q6tfni3b5a8	cl_1sizugbjggamtt	cl_pvlhgpc2mbanhq	md_89v98o1rpi4idv	cl_t4dwkmgadehh7w	cl_mhh2pb9uvhqu8d	\N	\N	\N	\N	2022-05-18 11:24:13.304236+02	2022-05-18 11:24:13.304236+02
ln_egnt3lpkoxkrnr	\N	hm	\N	\N	cl_n13v5fxz7h0dgn	md_cltahxtbesn8vs	cl_o1e6spelqpafwg	cl_ydjxav6o2pfyt2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.00818+02	2022-05-18 11:24:13.00818+02
ln_fj7zwumtipbs4a	\N	hm	\N	\N	cl_uqphjc9xpcdo96	md_yor91fufon7h6c	cl_tqy34owdpe210n	cl_99h5zfuj8v82zf	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.011984+02	2022-05-18 11:24:13.011984+02
ln_j8vslqs2ezg7nw	\N	bt	\N	\N	cl_q3y938j1f63exf	md_muq975426eve8k	cl_rfzkjum97sm4px	cl_zdgsrgtaf7i8jh	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.075478+02	2022-05-18 11:24:13.075478+02
ln_6eixvx6fxkx1zm	\N	bt	\N	\N	cl_ly7t4ezyypj5ld	md_gwspodfugzyzs2	cl_m5ihdmbagtnr1m	cl_yuc04l9uriu78e	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.105773+02	2022-05-18 11:24:13.105773+02
ln_ijeagtx54st25p	\N	bt	\N	\N	cl_e2mhvq3w5in28m	md_dg3rmvfk4bsg1i	cl_3gm2h79uxhdzcc	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.125549+02	2022-05-18 11:24:13.125549+02
ln_54h9jqps7njerf	\N	bt	\N	\N	cl_xdmlbpk1l1lxt6	md_dg3rmvfk4bsg1i	cl_c0p5yzzp5h2fdv	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.008429+02	2022-05-18 11:24:13.008429+02
ln_hmy20hpm9iv1wf	\N	bt	\N	\N	cl_74pt8m8ck2391p	md_47c5g91l1nc5se	cl_8vqk98wddq5tto	cl_da7co98ztmz4v6	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.012177+02	2022-05-18 11:24:13.012177+02
ln_1pzm2tkd5f5bse	\N	bt	\N	\N	cl_qdrhybis5cbz4b	md_ue8cshscw72tz4	cl_s38bav6sh7c4en	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.014419+02	2022-05-18 11:24:13.014419+02
ln_rpiudmuio137k6	\N	hm	\N	\N	cl_8cvwgubyxtqrep	md_5hjnvbjhgfiz6r	cl_cyde2oo9zyylv7	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.016314+02	2022-05-18 11:24:13.016314+02
ln_yv6zqavz7198hw	\N	bt	\N	\N	cl_il02qiiyaybb14	md_ofa8d1ocm4oowx	cl_bg6e57ixpkc66e	cl_eqywr4zelb2x2k	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.075836+02	2022-05-18 11:24:13.075836+02
ln_9mih85ti41eo4l	\N	hm	\N	\N	cl_kwcxljamy9qed3	md_ruvc4yz1y8yonz	cl_st7iuxws65atp5	cl_hl2a11v360e4b1	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.079291+02	2022-05-18 11:24:13.079291+02
ln_o0mn26ttzkhxym	\N	hm	\N	\N	cl_xnqmpmsx7xxspc	md_ebe9bibmq1l2uf	cl_actvro2vfybjau	cl_eqywr4zelb2x2k	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.106124+02	2022-05-18 11:24:13.106124+02
ln_0hmth8cv3mxkne	\N	bt	\N	\N	cl_z02hnsgf78dgxx	md_ajkugjxa4haf22	cl_2b7kd2rxaa6i09	cl_py299lng6pd9y4	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.008701+02	2022-05-18 11:24:13.008701+02
ln_cz1xbhir1h2rtf	\N	bt	\N	\N	cl_m4ojzvyzshvizr	md_vb4nqekkid2rl6	cl_48210hus0kkb2v	cl_bq67f15udwlp8d	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.012421+02	2022-05-18 11:24:13.012421+02
ln_und9unoumjoelt	\N	bt	\N	\N	cl_uule6x2nkxw84p	md_cltahxtbesn8vs	cl_w9k4oudkzve2tu	cl_hl2a11v360e4b1	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.014647+02	2022-05-18 11:24:13.014647+02
ln_uk7wea9e3ja23g	\N	bt	\N	\N	cl_xvb4bw8js0m928	md_ue8cshscw72tz4	cl_yaupebwbnv4qxt	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.016493+02	2022-05-18 11:24:13.016493+02
ln_ga3qtc52oixj19	\N	hm	\N	\N	cl_ed9em6bn75thju	md_cltahxtbesn8vs	cl_kmwlqydz7qisje	cl_eqywr4zelb2x2k	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.073467+02	2022-05-18 11:24:13.073467+02
ln_t6appp7jke6vu9	\N	bt	\N	\N	cl_ub0vo7g9n0zjrb	md_cltahxtbesn8vs	cl_fl7eap9v6vs6k5	cl_hl2a11v360e4b1	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.106311+02	2022-05-18 11:24:13.106311+02
ln_yrepq782a43rfp	\N	bt	\N	\N	cl_7s2mqpkdq9y4f5	md_ue8cshscw72tz4	cl_87e0ickdzoj44q	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.125844+02	2022-05-18 11:24:13.125844+02
ln_9d6f51ueqctbqy	\N	bt	\N	\N	cl_9xl9rhlhzm8ga2	md_ue8cshscw72tz4	cl_dp5y1a7pcz0s54	cl_mgbs2x2eusxqn9	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.142432+02	2022-05-18 11:24:13.142432+02
ln_oa2jsr8d9kclpj	\N	bt	\N	\N	cl_5u2tjgd4jktqj2	md_cltahxtbesn8vs	cl_st7iuxws65atp5	cl_hl2a11v360e4b1	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.154215+02	2022-05-18 11:24:13.154215+02
ln_fpwc6qlkw2ct2i	\N	bt	\N	\N	cl_hzx263qp16kmx1	md_ofa8d1ocm4oowx	cl_actvro2vfybjau	cl_eqywr4zelb2x2k	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.00989+02	2022-05-18 11:24:13.00989+02
ln_n6upc4sep6cmxo	\N	bt	\N	\N	cl_8jwbbixyl9l1gu	md_6ht8y7vgzcek0p	cl_t4dwkmgadehh7w	cl_1sizugbjggamtt	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.013261+02	2022-05-18 11:24:13.013261+02
ln_vh3n2dmgq2rcgo	\N	bt	\N	\N	cl_38nvbo6hofo1ba	md_ofa8d1ocm4oowx	cl_fikkm6avrnvalf	cl_eqywr4zelb2x2k	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.074277+02	2022-05-18 11:24:13.074277+02
ln_y60br1k0suv43b	\N	hm	\N	\N	cl_9zapplyeoqzc3w	md_r7cgbvzh9qy7qk	cl_2n98v3z6z9y6eo	cl_99h5zfuj8v82zf	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.077743+02	2022-05-18 11:24:13.077743+02
ln_6q2kqdn3y3c8yq	\N	bt	\N	\N	cl_l0q6rlkn5vn5ig	md_dg3rmvfk4bsg1i	cl_enjcwnqd9lm0hl	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.010454+02	2022-05-18 11:24:13.010454+02
ln_clrpju286ir2dk	\N	hm	\N	\N	cl_f3nxqy328zxk7v	md_47c5g91l1nc5se	cl_da7co98ztmz4v6	cl_hflevolpf863c2	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.074628+02	2022-05-18 11:24:13.074628+02
ln_1p0qcwbzs7q93l	\N	bt	\N	\N	cl_yl6pabbkd8b0yk	md_wjrbh5du4ifq5b	cl_lxtgajo4s3br49	cl_u7wce2nao3rabj	\N	\N	\N	NO ACTION	NO ACTION	\N	\N	2022-05-18 11:24:13.113851+02	2022-05-18 11:24:13.113851+02
\.


--
-- Data for Name: nc_col_rollup_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_col_rollup_v2 (id, fk_column_id, fk_relation_column_id, fk_rollup_column_id, rollup_function, deleted, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_col_select_options_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_col_select_options_v2 (id, fk_column_id, title, color, "order", created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_columns_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_columns_v2 (id, base_id, project_id, fk_model_id, title, column_name, uidt, dt, np, ns, clen, cop, pk, pv, rqd, un, ct, ai, "unique", cdf, cc, csn, dtx, dtxp, dtxs, au, validate, virtual, deleted, system, "order", created_at, updated_at) FROM stdin;
cl_jta9jl2qws2s1t	ds_aej9ao4mzw9cff	p_41297240e6of48	md_7vpsxklk3sph75	CustomerId	customer_id	ForeignKey	character	\N	\N	\N	1	t	t	t	f	\N	f	\N	\N	\N	\N	character	\N	\N	f	\N	\N	\N	t	1	2022-05-18 11:24:11.770813+02	2022-05-18 11:24:11.770813+02
cl_gpkukq2p7xpr9n	ds_aej9ao4mzw9cff	p_41297240e6of48	md_79w8haiknxbkzd	CategoryId	category_id	Number	smallint	16	0	\N	1	t	\N	t	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	f	1	2022-05-18 11:24:11.77178+02	2022-05-18 11:24:11.77178+02
cl_9wjh8ccwe2bpmz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_71j5r6adaq9l3i	CustomerTypeId	customer_type_id	SingleLineText	character	\N	\N	\N	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character	\N	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.77454+02	2022-05-18 11:24:11.77454+02
cl_bq67f15udwlp8d	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	CustomerId	customer_id	SingleLineText	character	\N	\N	\N	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character	\N	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.786278+02	2022-05-18 11:24:11.786278+02
cl_k5ansjw5dhsxmw	ds_aej9ao4mzw9cff	p_41297240e6of48	md_q2z56jvps22ukx	EmployeeId	employee_id	ForeignKey	smallint	16	0	\N	1	t	t	t	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	t	1	2022-05-18 11:24:11.793592+02	2022-05-18 11:24:11.793592+02
cl_t9q9o6itn46bdj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5hjnvbjhgfiz6r	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.80877+02	2022-05-18 11:24:11.80877+02
cl_mgbs2x2eusxqn9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.81167+02	2022-05-18 11:24:11.81167+02
cl_99h5zfuj8v82zf	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	EmployeeId	employee_id	Number	smallint	16	0	\N	1	t	\N	t	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	f	1	2022-05-18 11:24:11.813551+02	2022-05-18 11:24:11.813551+02
cl_ydjxav6o2pfyt2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.814453+02	2022-05-18 11:24:11.814453+02
cl_m85rflkcl4gkms	ds_aej9ao4mzw9cff	p_41297240e6of48	md_91fcdvmtveuz77	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.815026+02	2022-05-18 11:24:11.815026+02
cl_q8fyky7flb1b4w	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6pjpqivc2ymh3b	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.816005+02	2022-05-18 11:24:11.816005+02
cl_jccpvuyca95upl	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.816508+02	2022-05-18 11:24:11.816508+02
cl_zkzu34hln237a9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_tm1aur9f3yn2l2	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.81708+02	2022-05-18 11:24:11.81708+02
cl_pzpkxfvrsf6mbq	ds_aej9ao4mzw9cff	p_41297240e6of48	md_rmg6p4gh772ip8	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.818136+02	2022-05-18 11:24:11.818136+02
cl_5k0aee0pi9qjra	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.819256+02	2022-05-18 11:24:11.819256+02
cl_yuc04l9uriu78e	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.821099+02	2022-05-18 11:24:11.821099+02
cl_1947g4olf8l61e	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.822384+02	2022-05-18 11:24:11.822384+02
cl_haf6wz56kfhhoz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	BaseId	base_id	SingleLineText	character varying	\N	\N	20	1	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.824025+02	2022-05-18 11:24:11.824025+02
cl_1fvruzuc9dq17k	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.827349+02	2022-05-18 11:24:11.827349+02
cl_c7u4guohbetlzu	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	BaseId	base_id	SingleLineText	character varying	\N	\N	20	1	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.828318+02	2022-05-18 11:24:11.828318+02
cl_c0p5yzzp5h2fdv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_zmhiordb5vpwkf	FkViewId	fk_view_id	ForeignKey	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	1	2022-05-18 11:24:11.829513+02	2022-05-18 11:24:11.829513+02
cl_py299lng6pd9y4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	FkViewId	fk_view_id	ForeignKey	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	1	2022-05-18 11:24:11.830365+02	2022-05-18 11:24:11.830365+02
cl_gtefjqnxocsm95	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.83103+02	2022-05-18 11:24:11.83103+02
cl_7hk845mxxfw9nb	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.834217+02	2022-05-18 11:24:11.834217+02
cl_kd7e1qpculbik0	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.835371+02	2022-05-18 11:24:11.835371+02
cl_actvro2vfybjau	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	ProjectId	project_id	ForeignKey	character varying	\N	\N	128	1	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	t	1	2022-05-18 11:24:11.838689+02	2022-05-18 11:24:11.838689+02
cl_zdgsrgtaf7i8jh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.839096+02	2022-05-18 11:24:11.839096+02
cl_f1vep41nkme5bs	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.840153+02	2022-05-18 11:24:11.840153+02
cl_eqywr4zelb2x2k	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	Id	id	SingleLineText	character varying	\N	\N	128	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.843235+02	2022-05-18 11:24:11.843235+02
cl_k8wnnx3bnzcy95	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wqy1a8c5ggpyea	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.846734+02	2022-05-18 11:24:11.846734+02
cl_mhh2pb9uvhqu8d	ds_aej9ao4mzw9cff	p_41297240e6of48	md_89v98o1rpi4idv	OrgId	org_id	ForeignKey	character varying	\N	\N	20	1	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	1	2022-05-18 11:24:11.849032+02	2022-05-18 11:24:11.849032+02
cl_pvlhgpc2mbanhq	ds_aej9ao4mzw9cff	p_41297240e6of48	md_nj2q6tfni3b5a8	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.849572+02	2022-05-18 11:24:11.849572+02
cl_eztxs6bu5hddmh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_0z5ofu8o9xph1j	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.851063+02	2022-05-18 11:24:11.851063+02
cl_gnsoew7es07l63	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.85211+02	2022-05-18 11:24:11.85211+02
cl_hl2a11v360e4b1	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.852694+02	2022-05-18 11:24:11.852694+02
cl_duxrw3cakjkz2l	ds_aej9ao4mzw9cff	p_41297240e6of48	md_1wgrn82853mqsm	Id	id	Number	integer	32	0	\N	1	t	\N	t	f	\N	t	\N	nextval('nc_shared_bases_id_seq'::regclass)	\N	\N	integer	32	0	f	\N	\N	\N	f	1	2022-05-18 11:24:11.854462+02	2022-05-18 11:24:11.854462+02
cl_1sizugbjggamtt	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.863847+02	2022-05-18 11:24:11.863847+02
cl_lwxz8ubw0dlz2l	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	OrderId	order_id	Number	smallint	16	0	\N	1	t	\N	t	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	f	1	2022-05-18 11:24:11.868344+02	2022-05-18 11:24:11.868344+02
cl_68x1go22132wqm	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	ProductId	product_id	Number	smallint	16	0	\N	1	t	\N	t	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	f	1	2022-05-18 11:24:11.870949+02	2022-05-18 11:24:11.870949+02
cl_vvtlvtc6gzirw0	ds_aej9ao4mzw9cff	p_41297240e6of48	md_jgq6gbzvyfbk2g	RegionId	region_id	Number	smallint	16	0	\N	1	t	\N	t	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	f	1	2022-05-18 11:24:11.87135+02	2022-05-18 11:24:11.87135+02
cl_z4jql0g2ge0i9x	ds_aej9ao4mzw9cff	p_41297240e6of48	md_s9jny9njrhdnwz	OrderId	order_id	ForeignKey	smallint	16	0	\N	1	t	\N	t	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	t	1	2022-05-18 11:24:11.872216+02	2022-05-18 11:24:11.872216+02
cl_0qe00wlxhp6l8d	ds_aej9ao4mzw9cff	p_41297240e6of48	md_q2z56jvps22ukx	TerritoryId	territory_id	ForeignKey	character varying	\N	\N	20	2	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	2	2022-05-18 11:24:11.905815+02	2022-05-18 11:24:11.905815+02
cl_bg6e57ixpkc66e	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	ProjectId	project_id	ForeignKey	character varying	\N	\N	128	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	t	2	2022-05-18 11:24:11.922776+02	2022-05-18 11:24:11.922776+02
cl_9wmjbqjyqwqjfv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wqy1a8c5ggpyea	BaseId	base_id	SingleLineText	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.952324+02	2022-05-18 11:24:11.952324+02
cl_s7mtg7jpphuwfw	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wk0cg5v1lnk9ew	StateName	state_name	SingleLineText	character varying	\N	\N	100	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	100	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.982492+02	2022-05-18 11:24:11.982492+02
cl_8xhockmpbsi2md	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	Alias	alias	SingleLineText	character varying	\N	\N	255	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.032453+02	2022-05-18 11:24:12.032453+02
cl_ll4pwrwaq6kj4i	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	Heading	heading	SingleLineText	character varying	\N	\N	255	4	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.120713+02	2022-05-18 11:24:12.120713+02
cl_hffjtbw2nw06p1	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	Starred	starred	Checkbox	boolean	\N	\N	\N	4	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.140454+02	2022-05-18 11:24:12.140454+02
cl_6fchl3afgxh7yp	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	TitleOfCourtesy	title_of_courtesy	SingleLineText	character varying	\N	\N	25	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	25	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.198423+02	2022-05-18 11:24:12.198423+02
cl_db2hgifuodn7cs	ds_aej9ao4mzw9cff	p_41297240e6of48	md_zmhiordb5vpwkf	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	5	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.207082+02	2022-05-18 11:24:12.207082+02
cl_73k8fe81h06itk	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Title	title	SingleLineText	character varying	\N	\N	255	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.218958+02	2022-05-18 11:24:12.218958+02
cl_orefgn2y3dot57	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	Label	label	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.346537+02	2022-05-18 11:24:12.346537+02
cl_v0nkthutghwjgf	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Docs	docs	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.36321+02	2022-05-18 11:24:12.36321+02
cl_tux3x9lo761ow6	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	Help	help	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.419013+02	2022-05-18 11:24:12.419013+02
cl_ewcsxrn7zdfg6x	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	UnitsOnOrder	units_on_order	Number	smallint	16	0	\N	8	f	\N	f	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	f	8	2022-05-18 11:24:12.450589+02	2022-05-18 11:24:12.450589+02
cl_m2e1p9fzymlhq4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	Width	width	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	'200px'::character varying	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.478116+02	2022-05-18 11:24:12.478116+02
cl_8kcusz9r08t1n5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	SubmitAnotherForm	submit_another_form	Checkbox	boolean	\N	\N	\N	10	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.530653+02	2022-05-18 11:24:12.530653+02
cl_9dhvtte5npaada	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Logo	logo	SingleLineText	character varying	\N	\N	255	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.544499+02	2022-05-18 11:24:12.544499+02
cl_crrp1b4xmw6rye	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	11	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.584499+02	2022-05-18 11:24:12.584499+02
cl_fpfamco3hhd3sv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	11	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.59352+02	2022-05-18 11:24:12.59352+02
cl_pthbmkbtet9gao	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Cop	cop	SingleLineText	character varying	\N	\N	255	12	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.622083+02	2022-05-18 11:24:12.622083+02
cl_7uz6ir4vjgn1cx	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	13	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.664961+02	2022-05-18 11:24:12.664961+02
cl_rdcmlig4o9fv6y	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	Order	order	Decimal	real	24	\N	\N	13	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.676572+02	2022-05-18 11:24:12.676572+02
cl_ufirlt09c46y33	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	ExecutionTime	execution_time	Number	integer	32	0	\N	15	f	\N	f	f	\N	f	\N	\N	\N	\N	integer	32	0	f	\N	\N	\N	f	15	2022-05-18 11:24:12.719483+02	2022-05-18 11:24:12.719483+02
cl_kmwn5f21u4wyxf	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	Roles	roles	SingleLineText	character varying	\N	\N	255	15	f	\N	f	f	\N	f	\N	'editor'::character varying	\N	\N	character varying	255	\N	f	\N	\N	\N	f	15	2022-05-18 11:24:12.729804+02	2022-05-18 11:24:12.729804+02
cl_66atpjp42alzgd	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	16	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	16	2022-05-18 11:24:12.740162+02	2022-05-18 11:24:12.740162+02
cl_yqr67t870n3viz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Ct	ct	LongText	text	\N	\N	\N	17	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	17	2022-05-18 11:24:12.754759+02	2022-05-18 11:24:12.754759+02
cl_vj4rthimqxnzck	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Ai	ai	Checkbox	boolean	\N	\N	\N	18	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	18	2022-05-18 11:24:12.771098+02	2022-05-18 11:24:12.771098+02
cl_ltyxan2j8mc59f	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Unique	unique	Checkbox	boolean	\N	\N	\N	19	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	19	2022-05-18 11:24:12.784894+02	2022-05-18 11:24:12.784894+02
cl_3s6cb84a61zbtv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Cdf	cdf	LongText	text	\N	\N	\N	20	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	20	2022-05-18 11:24:12.791691+02	2022-05-18 11:24:12.791691+02
cl_s5uljkpg4meb33	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Cc	cc	LongText	text	\N	\N	\N	21	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	21	2022-05-18 11:24:12.797136+02	2022-05-18 11:24:12.797136+02
cl_yr12t2sier2m7f	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Csn	csn	SingleLineText	character varying	\N	\N	255	22	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	22	2022-05-18 11:24:12.801267+02	2022-05-18 11:24:12.801267+02
cl_flrhvi2qd887s8	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Dtx	dtx	SingleLineText	character varying	\N	\N	255	23	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	23	2022-05-18 11:24:12.805221+02	2022-05-18 11:24:12.805221+02
cl_r2anw6zpujmv39	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Dtxp	dtxp	LongText	text	\N	\N	\N	24	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	24	2022-05-18 11:24:12.808339+02	2022-05-18 11:24:12.808339+02
cl_n662hucflgmu9t	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Dtxs	dtxs	SingleLineText	character varying	\N	\N	255	25	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	25	2022-05-18 11:24:12.813325+02	2022-05-18 11:24:12.813325+02
cl_mzo2rej6vzbu49	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Au	au	Checkbox	boolean	\N	\N	\N	26	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	26	2022-05-18 11:24:12.817322+02	2022-05-18 11:24:12.817322+02
cl_r6h0z44rb8rmgu	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Validate	validate	LongText	text	\N	\N	\N	27	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	27	2022-05-18 11:24:12.821176+02	2022-05-18 11:24:12.821176+02
cl_mkav6fudiwsrvy	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Virtual	virtual	Checkbox	boolean	\N	\N	\N	28	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	28	2022-05-18 11:24:12.824502+02	2022-05-18 11:24:12.824502+02
cl_r7slbxotamjxlq	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Deleted	deleted	Checkbox	boolean	\N	\N	\N	29	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	29	2022-05-18 11:24:12.829358+02	2022-05-18 11:24:12.829358+02
cl_hmy4b3d07hbwwm	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	System	system	Checkbox	boolean	\N	\N	\N	30	f	\N	f	f	\N	f	\N	false	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	30	2022-05-18 11:24:12.832992+02	2022-05-18 11:24:12.832992+02
cl_ji20zt281pec2c	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	SupplierId	supplier_id	Number	smallint	16	0	\N	1	t	\N	t	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	f	1	2022-05-18 11:24:11.873497+02	2022-05-18 11:24:11.873497+02
cl_px1s562w4pnn24	ds_aej9ao4mzw9cff	p_41297240e6of48	md_71j5r6adaq9l3i	CustomerDesc	customer_desc	LongText	text	\N	\N	\N	2	f	t	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.896314+02	2022-05-18 11:24:11.896314+02
cl_narv6h4qk6asbn	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6pjpqivc2ymh3b	FkColumnId	fk_column_id	ForeignKey	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	2	2022-05-18 11:24:11.92423+02	2022-05-18 11:24:11.92423+02
cl_291qx9ou1akw8y	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	User	user	SingleLineText	character varying	\N	\N	255	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.929882+02	2022-05-18 11:24:11.929882+02
cl_swryd2y3ho6c3t	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	2	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.933896+02	2022-05-18 11:24:11.933896+02
cl_48210hus0kkb2v	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	CustomerId	customer_id	ForeignKey	character	\N	\N	\N	2	f	t	f	f	\N	f	\N	\N	\N	\N	character	\N	\N	f	\N	\N	\N	t	2	2022-05-18 11:24:11.970915+02	2022-05-18 11:24:11.970915+02
cl_cwask6ptfuopet	ds_aej9ao4mzw9cff	p_41297240e6of48	md_jgq6gbzvyfbk2g	RegionDescription	region_description	SingleLineText	character	\N	\N	\N	2	f	t	t	f	\N	f	\N	\N	\N	\N	character	\N	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.975343+02	2022-05-18 11:24:11.975343+02
cl_vn7qtpv2s09tkt	ds_aej9ao4mzw9cff	p_41297240e6of48	md_79w8haiknxbkzd	Description	description	LongText	text	\N	\N	\N	3	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:11.989741+02	2022-05-18 11:24:11.989741+02
cl_da7co98ztmz4v6	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	FkViewId	fk_view_id	ForeignKey	character varying	\N	\N	20	3	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	3	2022-05-18 11:24:12.034321+02	2022-05-18 11:24:12.034321+02
cl_u4mw1fxp0si97k	ds_aej9ao4mzw9cff	p_41297240e6of48	md_zmhiordb5vpwkf	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.040756+02	2022-05-18 11:24:12.040756+02
cl_aemppv0wnynryh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	Prefix	prefix	SingleLineText	character varying	\N	\N	255	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.05373+02	2022-05-18 11:24:12.05373+02
cl_wn3aloaldty24k	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	FkModelId	fk_model_id	ForeignKey	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	4	2022-05-18 11:24:12.110668+02	2022-05-18 11:24:12.110668+02
cl_ytfwqrsq3akllw	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6pjpqivc2ymh3b	Color	color	SingleLineText	character varying	\N	\N	255	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.115651+02	2022-05-18 11:24:12.115651+02
cl_luguf02s068frs	ds_aej9ao4mzw9cff	p_41297240e6of48	md_rmg6p4gh772ip8	FkRollupColumnId	fk_rollup_column_id	ForeignKey	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	4	2022-05-18 11:24:12.119006+02	2022-05-18 11:24:12.119006+02
cl_k5fdsl202of2rn	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	Show	show	Checkbox	boolean	\N	\N	\N	4	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.12984+02	2022-05-18 11:24:12.12984+02
cl_l08wx8t5fcd9lu	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Active	active	Checkbox	boolean	\N	\N	\N	4	f	\N	f	f	\N	f	\N	false	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.134703+02	2022-05-18 11:24:12.134703+02
cl_ove7ejfffgx03c	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	Salt	salt	SingleLineText	character varying	\N	\N	255	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.155601+02	2022-05-18 11:24:12.155601+02
cl_42mxp318j7y76i	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	OrderDate	order_date	Date	date	\N	\N	\N	4	f	\N	f	f	\N	f	\N	\N	\N	\N	date	0	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.158553+02	2022-05-18 11:24:12.158553+02
cl_gi9vm39qv3xdin	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	ContactTitle	contact_title	SingleLineText	character varying	\N	\N	30	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	30	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.170279+02	2022-05-18 11:24:12.170279+02
cl_rfzkjum97sm4px	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	FkHookId	fk_hook_id	ForeignKey	character varying	\N	\N	20	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	5	2022-05-18 11:24:12.20282+02	2022-05-18 11:24:12.20282+02
cl_9a6x2krsshjkks	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	Order	order	Decimal	real	24	\N	\N	5	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.208081+02	2022-05-18 11:24:12.208081+02
cl_k4waejjah5twfj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_1wgrn82853mqsm	SharedBaseId	shared_base_id	SingleLineText	character varying	\N	\N	255	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.222956+02	2022-05-18 11:24:12.222956+02
cl_v63cf04xvz97pb	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	FkColumnId	fk_column_id	ForeignKey	character varying	\N	\N	20	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	6	2022-05-18 11:24:12.272864+02	2022-05-18 11:24:12.272864+02
cl_e98v7uxkxggrni	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	ShowAllFields	show_all_fields	Checkbox	boolean	\N	\N	\N	6	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.292265+02	2022-05-18 11:24:12.292265+02
cl_zpybi9h1u394ao	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	RowId	row_id	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.337517+02	2022-05-18 11:24:12.337517+02
cl_p8rug1g5fmx2rz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Uidt	uidt	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.341869+02	2022-05-18 11:24:12.341869+02
cl_s4dj2g6xpllcm8	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	AllowCopy	allow_copy	Checkbox	boolean	\N	\N	\N	7	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.373907+02	2022-05-18 11:24:12.373907+02
cl_lxtgajo4s3br49	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	ShipVia	ship_via	ForeignKey	smallint	16	0	\N	7	f	\N	f	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	t	7	2022-05-18 11:24:12.38539+02	2022-05-18 11:24:12.38539+02
cl_z7tkdlhl3cxjj9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Dt	dt	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.420198+02	2022-05-18 11:24:12.420198+02
cl_u1ygzte476l7jn	ds_aej9ao4mzw9cff	p_41297240e6of48	md_1wgrn82853mqsm	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	8	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.440271+02	2022-05-18 11:24:12.440271+02
cl_dp2o9412pva5ph	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	Show	show	Checkbox	boolean	\N	\N	\N	9	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.491878+02	2022-05-18 11:24:12.491878+02
cl_sf0a0bdn3w6jfi	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	Password	password	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.496764+02	2022-05-18 11:24:12.496764+02
cl_a00e6lm16vrvm9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	InviteToken	invite_token	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.503209+02	2022-05-18 11:24:12.503209+02
cl_rfdwwrsb228fls	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	Region	region	SingleLineText	character varying	\N	\N	15	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	15	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.531864+02	2022-05-18 11:24:12.531864+02
cl_0biojey4j3nxim	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	Value	value	SingleLineText	character varying	\N	\N	255	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.535061+02	2022-05-18 11:24:12.535061+02
cl_fl7eap9v6vs6k5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	FkMmModelId	fk_mm_model_id	ForeignKey	character varying	\N	\N	20	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	10	2022-05-18 11:24:12.538327+02	2022-05-18 11:24:12.538327+02
cl_rvh480evb8owh2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	InviteTokenExpires	invite_token_expires	SingleLineText	character varying	\N	\N	255	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.553793+02	2022-05-18 11:24:12.553793+02
cl_wsrry9ph73bbf9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Async	async	Checkbox	boolean	\N	\N	\N	11	f	\N	f	f	\N	f	\N	false	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.591105+02	2022-05-18 11:24:12.591105+02
cl_rg5pf6ur5s8ofd	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	Tags	tags	SingleLineText	character varying	\N	\N	255	11	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.595999+02	2022-05-18 11:24:12.595999+02
cl_hflevolpf863c2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	Id	id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.874509+02	2022-05-18 11:24:11.874509+02
cl_6i16gsjxn2t5bh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_j409llgiqo90yz	TerritoryId	territory_id	SingleLineText	character varying	\N	\N	20	1	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	1	2022-05-18 11:24:11.87822+02	2022-05-18 11:24:11.87822+02
cl_e45d0ubf4cepmg	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	BaseId	base_id	SingleLineText	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.934617+02	2022-05-18 11:24:11.934617+02
cl_t4dwkmgadehh7w	ds_aej9ao4mzw9cff	p_41297240e6of48	md_89v98o1rpi4idv	UserId	user_id	ForeignKey	character varying	\N	\N	20	2	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	2	2022-05-18 11:24:11.958572+02	2022-05-18 11:24:11.958572+02
cl_u6d2985s4k2vje	ds_aej9ao4mzw9cff	p_41297240e6of48	md_s9jny9njrhdnwz	ProductId	product_id	ForeignKey	smallint	16	0	\N	2	t	\N	t	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	t	2	2022-05-18 11:24:11.980964+02	2022-05-18 11:24:11.980964+02
cl_mglgw6q3hpyxw5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	ContactName	contact_name	SingleLineText	character varying	\N	\N	30	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	30	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.001419+02	2022-05-18 11:24:12.001419+02
cl_lbtwbffy5ca4x5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6pjpqivc2ymh3b	Title	title	SingleLineText	character varying	\N	\N	255	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.024373+02	2022-05-18 11:24:12.024373+02
cl_5gkj8u9n2irl5t	ds_aej9ao4mzw9cff	p_41297240e6of48	md_rmg6p4gh772ip8	FkRelationColumnId	fk_relation_column_id	ForeignKey	character varying	\N	\N	20	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	3	2022-05-18 11:24:12.030779+02	2022-05-18 11:24:12.030779+02
cl_us0qzk2ufx407a	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.03492+02	2022-05-18 11:24:12.03492+02
cl_isb3w8xrlllgwm	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.046707+02	2022-05-18 11:24:12.046707+02
cl_0j274wxlvyo4yz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	Roles	roles	LongText	text	\N	\N	\N	3	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.051439+02	2022-05-18 11:24:12.051439+02
cl_7fuircxtsky4ng	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wqy1a8c5ggpyea	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.054525+02	2022-05-18 11:24:12.054525+02
cl_fwqr5zoahf4y1d	ds_aej9ao4mzw9cff	p_41297240e6of48	md_nj2q6tfni3b5a8	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	3	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.0577+02	2022-05-18 11:24:12.0577+02
cl_pwh6s0lomxswl7	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	SupplierId	supplier_id	ForeignKey	smallint	16	0	\N	3	f	\N	f	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	t	3	2022-05-18 11:24:12.074813+02	2022-05-18 11:24:12.074813+02
cl_hi6bftt57005lu	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	ContactTitle	contact_title	SingleLineText	character varying	\N	\N	30	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	30	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.097973+02	2022-05-18 11:24:12.097973+02
cl_8vqk98wddq5tto	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	FkViewId	fk_view_id	ForeignKey	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	4	2022-05-18 11:24:12.119493+02	2022-05-18 11:24:12.119493+02
cl_nb9fflwnh0el52	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	Status	status	SingleLineText	character varying	\N	\N	255	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.142212+02	2022-05-18 11:24:12.142212+02
cl_2l1uwu50cen5px	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	FkModelId	fk_model_id	ForeignKey	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	4	2022-05-18 11:24:12.159284+02	2022-05-18 11:24:12.159284+02
cl_rk5bzouixq4oyt	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	CategoryId	category_id	ForeignKey	smallint	16	0	\N	4	f	\N	f	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	t	4	2022-05-18 11:24:12.16464+02	2022-05-18 11:24:12.16464+02
cl_i2d4othvsshdrf	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6pjpqivc2ymh3b	Order	order	Decimal	real	24	\N	\N	5	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.19676+02	2022-05-18 11:24:12.19676+02
cl_fikkm6avrnvalf	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	ProjectId	project_id	ForeignKey	character varying	\N	\N	128	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	t	5	2022-05-18 11:24:12.200303+02	2022-05-18 11:24:12.200303+02
cl_6yy0dvgw2yb248	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	FkColumnId	fk_column_id	ForeignKey	character varying	\N	\N	20	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	5	2022-05-18 11:24:12.203399+02	2022-05-18 11:24:12.203399+02
cl_3kjqc5f1aywd4g	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	DbType	db_type	SingleLineText	character varying	\N	\N	255	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.20578+02	2022-05-18 11:24:12.20578+02
cl_cdefrh9aw9kurm	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	PrevEnabled	prev_enabled	Checkbox	boolean	\N	\N	\N	5	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.213522+02	2022-05-18 11:24:12.213522+02
cl_yws811j3krz71y	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	Pinned	pinned	Checkbox	boolean	\N	\N	\N	5	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.220501+02	2022-05-18 11:24:12.220501+02
cl_h0pdckpzut20m0	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	RequiredDate	required_date	Date	date	\N	\N	\N	5	f	\N	f	f	\N	f	\N	\N	\N	\N	date	0	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.236423+02	2022-05-18 11:24:12.236423+02
cl_ulp62fxwdxscyv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	Title	title	SingleLineText	character varying	\N	\N	255	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.239311+02	2022-05-18 11:24:12.239311+02
cl_t92sq868h7db4w	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	FkModelId	fk_model_id	ForeignKey	character varying	\N	\N	20	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	6	2022-05-18 11:24:12.270567+02	2022-05-18 11:24:12.270567+02
cl_6zmi1axycbm9zs	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	Uuid	uuid	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.273453+02	2022-05-18 11:24:12.273453+02
cl_lgjbx0b7bhizkc	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	Meta	meta	LongText	text	\N	\N	\N	6	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.298643+02	2022-05-18 11:24:12.298643+02
cl_edzzu8r2caweg4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	UnitPrice	unit_price	Decimal	real	24	\N	\N	6	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.30979+02	2022-05-18 11:24:12.30979+02
cl_lnb2rstu0dzd0s	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5hjnvbjhgfiz6r	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	7	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.342951+02	2022-05-18 11:24:12.342951+02
cl_pluaphsmwwqq5i	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	RedirectUrl	redirect_url	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.349334+02	2022-05-18 11:24:12.349334+02
cl_yaupebwbnv4qxt	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	FkCoverImageColId	fk_cover_image_col_id	ForeignKey	character varying	\N	\N	20	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	7	2022-05-18 11:24:12.3544+02	2022-05-18 11:24:12.3544+02
cl_biry9fx9h2mhne	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	Color	color	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.374524+02	2022-05-18 11:24:12.374524+02
cl_jufntqmsiohfs3	ds_aej9ao4mzw9cff	p_41297240e6of48	md_1wgrn82853mqsm	Password	password	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.379202+02	2022-05-18 11:24:12.379202+02
cl_m6xqxi37t6xbx8	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	IsDefault	is_default	Checkbox	boolean	\N	\N	\N	7	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.388731+02	2022-05-18 11:24:12.388731+02
cl_r4ephzb7n37bj1	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	RedirectAfterSecs	redirect_after_secs	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.420808+02	2022-05-18 11:24:12.420808+02
cl_mk3hu3qmq6tdfj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	Help	help	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.433055+02	2022-05-18 11:24:12.433055+02
cl_0x0h3rq3ygplwc	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wk0cg5v1lnk9ew	StateId	state_id	Number	smallint	16	0	\N	1	t	\N	t	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	f	1	2022-05-18 11:24:11.877155+02	2022-05-18 11:24:11.877155+02
cl_7vy94bkvan3b13	ds_aej9ao4mzw9cff	p_41297240e6of48	md_7vpsxklk3sph75	CustomerTypeId	customer_type_id	ForeignKey	character	\N	\N	\N	2	t	\N	t	f	\N	f	\N	\N	\N	\N	character	\N	\N	f	\N	\N	\N	t	2	2022-05-18 11:24:11.889062+02	2022-05-18 11:24:11.889062+02
cl_2xu900p2fe407f	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	BaseId	base_id	SingleLineText	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.920861+02	2022-05-18 11:24:11.920861+02
cl_ycnsnkvz4v8akf	ds_aej9ao4mzw9cff	p_41297240e6of48	md_tm1aur9f3yn2l2	FkColumnId	fk_column_id	ForeignKey	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	2	2022-05-18 11:24:11.924835+02	2022-05-18 11:24:11.924835+02
cl_dpxgtdswb10tkh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_rmg6p4gh772ip8	FkColumnId	fk_column_id	ForeignKey	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	2	2022-05-18 11:24:11.930302+02	2022-05-18 11:24:11.930302+02
cl_tgvxn298rrqb61	ds_aej9ao4mzw9cff	p_41297240e6of48	md_zmhiordb5vpwkf	BaseId	base_id	SingleLineText	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.941055+02	2022-05-18 11:24:11.941055+02
cl_rbx0v8hbfg9saz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	Email	email	SingleLineText	character varying	\N	\N	255	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.967309+02	2022-05-18 11:24:11.967309+02
cl_m1g535oorspd4b	ds_aej9ao4mzw9cff	p_41297240e6of48	md_llr0f8esceyorr	Name	name	SingleLineText	character varying	\N	\N	255	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.984421+02	2022-05-18 11:24:11.984421+02
cl_oq9xmxjup9n57n	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wjrbh5du4ifq5b	CompanyName	company_name	SingleLineText	character varying	\N	\N	40	2	f	t	t	f	\N	f	\N	\N	\N	\N	character varying	40	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.987107+02	2022-05-18 11:24:11.987107+02
cl_o5rwys2v3e6oyo	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	Ip	ip	SingleLineText	character varying	\N	\N	255	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.03035+02	2022-05-18 11:24:12.03035+02
cl_v28tzp0c0c0jt4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.037839+02	2022-05-18 11:24:12.037839+02
cl_da3hlyk05rx1wf	ds_aej9ao4mzw9cff	p_41297240e6of48	md_1wgrn82853mqsm	DbAlias	db_alias	SingleLineText	character varying	\N	\N	255	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.061832+02	2022-05-18 11:24:12.061832+02
cl_02u9gqs7dleiah	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.074453+02	2022-05-18 11:24:12.074453+02
cl_ep1ppfh31d8q0b	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	ContactName	contact_name	SingleLineText	character varying	\N	\N	30	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	30	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.080168+02	2022-05-18 11:24:12.080168+02
cl_c548hgzjy1v6kv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_79w8haiknxbkzd	Picture	picture	SpecificDBType	bytea	\N	\N	\N	4	f	\N	f	f	\N	f	\N	\N	\N	\N	bytea	\N	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.083991+02	2022-05-18 11:24:12.083991+02
cl_vhim94i6ia8ld4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_tm1aur9f3yn2l2	FormulaRaw	formula_raw	LongText	text	\N	\N	\N	4	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.115997+02	2022-05-18 11:24:12.115997+02
cl_sf4me3yekzw8kw	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	Virtual	virtual	Checkbox	boolean	\N	\N	\N	4	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.124677+02	2022-05-18 11:24:12.124677+02
cl_2b7kd2rxaa6i09	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	FkViewId	fk_view_id	ForeignKey	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	4	2022-05-18 11:24:12.138627+02	2022-05-18 11:24:12.138627+02
cl_j56fpioga326or	ds_aej9ao4mzw9cff	p_41297240e6of48	md_89v98o1rpi4idv	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	4	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.141765+02	2022-05-18 11:24:12.141765+02
cl_nv1tein75ejh1j	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	QuantityPerUnit	quantity_per_unit	SingleLineText	character varying	\N	\N	20	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.241857+02	2022-05-18 11:24:12.241857+02
cl_s38bav6sh7c4en	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	FkColumnId	fk_column_id	ForeignKey	character varying	\N	\N	20	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	6	2022-05-18 11:24:12.276348+02	2022-05-18 11:24:12.276348+02
cl_e72jcq87j4l87q	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Description	description	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.289294+02	2022-05-18 11:24:12.289294+02
cl_nwguwudzn15f31	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6pjpqivc2ymh3b	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	7	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.337925+02	2022-05-18 11:24:12.337925+02
cl_pasy3zcvs1sdoi	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	Title	title	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.353717+02	2022-05-18 11:24:12.353717+02
cl_8v2ynimciarahr	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wqy1a8c5ggpyea	Order	order	Decimal	real	24	\N	\N	7	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.359024+02	2022-05-18 11:24:12.359024+02
cl_qzvzbr3q5bxe4f	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Env	env	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	'all'::character varying	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.370263+02	2022-05-18 11:24:12.370263+02
cl_d8isivipdmp7ka	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	Uuid	uuid	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.437029+02	2022-05-18 11:24:12.437029+02
cl_iggt7rotz7o8fe	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	Email	email	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.48088+02	2022-05-18 11:24:12.48088+02
cl_sc5t1gtu5agi6b	ds_aej9ao4mzw9cff	p_41297240e6of48	md_tm1aur9f3yn2l2	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	9	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.486657+02	2022-05-18 11:24:12.486657+02
cl_js7114p3q9t5sa	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	Show	show	Checkbox	boolean	\N	\N	\N	9	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.492421+02	2022-05-18 11:24:12.492421+02
cl_9sqwmb4k4dxhoc	ds_aej9ao4mzw9cff	p_41297240e6of48	md_1wgrn82853mqsm	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	9	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.497066+02	2022-05-18 11:24:12.497066+02
cl_0i5dworwktqo5u	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	Required	required	Checkbox	boolean	\N	\N	\N	10	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.532378+02	2022-05-18 11:24:12.532378+02
cl_fv2czdtf2qgrw6	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Operation	operation	SingleLineText	character varying	\N	\N	255	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.54745+02	2022-05-18 11:24:12.54745+02
cl_9sh3add4j1h5tu	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	Phone	phone	SingleLineText	character varying	\N	\N	24	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	24	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.562052+02	2022-05-18 11:24:12.562052+02
cl_uqz0n2ikliepax	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	Order	order	Decimal	real	24	\N	\N	11	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.580012+02	2022-05-18 11:24:12.580012+02
cl_ygqq2ohbanufuo	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	Show	show	Checkbox	boolean	\N	\N	\N	11	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.583228+02	2022-05-18 11:24:12.583228+02
cl_z86x61x8bf46a2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	Notification	notification	LongText	text	\N	\N	\N	11	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.585908+02	2022-05-18 11:24:12.585908+02
cl_bavibzttofdtq2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	11	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.591563+02	2022-05-18 11:24:12.591563+02
cl_pnacxw8tomau6f	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	12	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.633706+02	2022-05-18 11:24:12.633706+02
cl_u7wce2nao3rabj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wjrbh5du4ifq5b	ShipperId	shipper_id	Number	smallint	16	0	\N	1	t	\N	t	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	f	1	2022-05-18 11:24:11.880177+02	2022-05-18 11:24:11.880177+02
cl_iajc7q7keo69iv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	BaseId	base_id	SingleLineText	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.931612+02	2022-05-18 11:24:11.931612+02
cl_iwu0vwdgt2ax7s	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	BaseId	base_id	SingleLineText	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.942056+02	2022-05-18 11:24:11.942056+02
cl_xzhpoilm6b0bmh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_1wgrn82853mqsm	ProjectId	project_id	SingleLineText	character varying	\N	\N	255	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.964534+02	2022-05-18 11:24:11.964534+02
cl_8rpusw386qct5p	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.031451+02	2022-05-18 11:24:12.031451+02
cl_bkvskjvf9kp0sk	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	Type	type	SingleLineText	character varying	\N	\N	255	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.035679+02	2022-05-18 11:24:12.035679+02
cl_xj1mgb1h75y03n	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Description	description	LongText	text	\N	\N	\N	3	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.047633+02	2022-05-18 11:24:12.047633+02
cl_2va4hiqfh4srjj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.051951+02	2022-05-18 11:24:12.051951+02
cl_h2is03dfyemdhi	ds_aej9ao4mzw9cff	p_41297240e6of48	md_0z5ofu8o9xph1j	OrgId	org_id	ForeignKey	character varying	\N	\N	20	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	3	2022-05-18 11:24:12.055276+02	2022-05-18 11:24:12.055276+02
cl_b4w9b6bqoq0mj1	ds_aej9ao4mzw9cff	p_41297240e6of48	md_llr0f8esceyorr	Batch	batch	Number	integer	32	0	\N	3	f	\N	f	f	\N	f	\N	\N	\N	\N	integer	32	0	f	\N	\N	\N	f	3	2022-05-18 11:24:12.081477+02	2022-05-18 11:24:12.081477+02
cl_roh6z8p6y9kt74	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	Config	config	LongText	text	\N	\N	\N	4	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.122893+02	2022-05-18 11:24:12.122893+02
cl_jprcavrmue3v7e	ds_aej9ao4mzw9cff	p_41297240e6of48	md_91fcdvmtveuz77	FkViewId	fk_view_id	ForeignKey	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	4	2022-05-18 11:24:12.12612+02	2022-05-18 11:24:12.12612+02
cl_yf33uxsgmmtxzo	ds_aej9ao4mzw9cff	p_41297240e6of48	md_nj2q6tfni3b5a8	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	4	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.143478+02	2022-05-18 11:24:12.143478+02
cl_s1ik0ltcq5ufcv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_1wgrn82853mqsm	Roles	roles	SingleLineText	character varying	\N	\N	255	4	f	\N	f	f	\N	f	\N	'viewer'::character varying	\N	\N	character varying	255	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.148152+02	2022-05-18 11:24:12.148152+02
cl_z04rfjpq8ajuaj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_s9jny9njrhdnwz	Quantity	quantity	Number	smallint	16	0	\N	4	f	\N	t	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	f	4	2022-05-18 11:24:12.165371+02	2022-05-18 11:24:12.165371+02
cl_0pok7tlgudc71r	ds_aej9ao4mzw9cff	p_41297240e6of48	md_llr0f8esceyorr	MigrationTime	migration_time	DateTime	timestamp with time zone	\N	\N	\N	4	f	\N	f	f	\N	f	\N	\N	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.168697+02	2022-05-18 11:24:12.168697+02
cl_swwufmsw23vux6	ds_aej9ao4mzw9cff	p_41297240e6of48	md_tm1aur9f3yn2l2	Error	error	LongText	text	\N	\N	\N	5	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.201024+02	2022-05-18 11:24:12.201024+02
cl_nrckcor6haqx4w	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	Meta	meta	LongText	text	\N	\N	\N	5	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.203941+02	2022-05-18 11:24:12.203941+02
cl_cz33iocpqead80	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.206445+02	2022-05-18 11:24:12.206445+02
cl_pi3yol2gqnb28k	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wqy1a8c5ggpyea	FkColumnId	fk_column_id	ForeignKey	character varying	\N	\N	20	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	5	2022-05-18 11:24:12.217927+02	2022-05-18 11:24:12.217927+02
cl_1qvn468mxp9xd0	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	ViewId	view_id	SingleLineText	character varying	\N	\N	255	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.22126+02	2022-05-18 11:24:12.22126+02
cl_s04j0dacnnx6jh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_0z5ofu8o9xph1j	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	5	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.224415+02	2022-05-18 11:24:12.224415+02
cl_b8pnbjqcsfv8yc	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	City	city	SingleLineText	character varying	\N	\N	15	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	15	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.252888+02	2022-05-18 11:24:12.252888+02
cl_hntenvlef6j2xi	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6pjpqivc2ymh3b	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	6	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.267926+02	2022-05-18 11:24:12.267926+02
cl_eid5i5cfqi101t	ds_aej9ao4mzw9cff	p_41297240e6of48	md_rmg6p4gh772ip8	Deleted	deleted	Checkbox	boolean	\N	\N	\N	6	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.27411+02	2022-05-18 11:24:12.27411+02
cl_wyit1f1uhnr4ux	ds_aej9ao4mzw9cff	p_41297240e6of48	md_91fcdvmtveuz77	Disabled	disabled	Checkbox	boolean	\N	\N	\N	6	f	\N	f	f	\N	f	\N	true	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.278613+02	2022-05-18 11:24:12.278613+02
cl_cbwd42kx2q4778	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	IsMeta	is_meta	Checkbox	boolean	\N	\N	\N	6	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.282439+02	2022-05-18 11:24:12.282439+02
cl_o4gbtsb3rs0i4x	ds_aej9ao4mzw9cff	p_41297240e6of48	md_zmhiordb5vpwkf	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	6	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.286805+02	2022-05-18 11:24:12.286805+02
cl_ocpkpwek7kd96e	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	Uuid	uuid	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.293745+02	2022-05-18 11:24:12.293745+02
cl_min1a7d8k76n5i	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	ShippedDate	shipped_date	Date	date	\N	\N	\N	6	f	\N	f	f	\N	f	\N	\N	\N	\N	date	0	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.305038+02	2022-05-18 11:24:12.305038+02
cl_3a45lb54s4im5e	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	Type	type	Number	integer	32	0	\N	6	f	\N	f	f	\N	f	\N	\N	\N	\N	integer	32	0	f	\N	\N	\N	f	6	2022-05-18 11:24:12.307788+02	2022-05-18 11:24:12.307788+02
cl_m5ihdmbagtnr1m	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	FkParentId	fk_parent_id	ForeignKey	character varying	\N	\N	20	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	7	2022-05-18 11:24:12.344741+02	2022-05-18 11:24:12.344741+02
cl_v0sietsica1ttd	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	Label	label	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.371585+02	2022-05-18 11:24:12.371585+02
cl_cvotd774eyvz5m	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	Username	username	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.380349+02	2022-05-18 11:24:12.380349+02
cl_hmh16s2ljlm5rq	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	Region	region	SingleLineText	character varying	\N	\N	15	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	15	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.392239+02	2022-05-18 11:24:12.392239+02
cl_c6hsuzyhqhcejg	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	Address	address	SingleLineText	character varying	\N	\N	60	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	60	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.418123+02	2022-05-18 11:24:12.418123+02
cl_x1knxeqizscbih	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	CoverImage	cover_image	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.424632+02	2022-05-18 11:24:12.424632+02
cl_vfsxjg732tgz6a	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	City	city	SingleLineText	character varying	\N	\N	15	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	15	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.476605+02	2022-05-18 11:24:12.476605+02
cl_m2mfi2laj0zcuv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	ComparisonOp	comparison_op	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.481521+02	2022-05-18 11:24:12.481521+02
cl_51c81maycrxne7	ds_aej9ao4mzw9cff	p_41297240e6of48	md_llr0f8esceyorr	Id	id	Number	integer	32	0	\N	1	t	\N	t	f	\N	t	\N	nextval('xc_knex_migrationsv2_id_seq'::regclass)	\N	\N	integer	32	0	f	\N	\N	\N	f	1	2022-05-18 11:24:11.88082+02	2022-05-18 11:24:11.88082+02
cl_dvr9xeexoglp9b	ds_aej9ao4mzw9cff	p_41297240e6of48	md_c1y2qwn3okt75w	Index	index	Number	integer	32	0	\N	1	t	\N	t	f	\N	t	\N	nextval('xc_knex_migrationsv2_lock_index_seq'::regclass)	\N	\N	integer	32	0	f	\N	\N	\N	f	1	2022-05-18 11:24:11.887471+02	2022-05-18 11:24:11.887471+02
cl_cyde2oo9zyylv7	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5hjnvbjhgfiz6r	FkColumnId	fk_column_id	ForeignKey	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	2	2022-05-18 11:24:11.919051+02	2022-05-18 11:24:11.919051+02
cl_p7fzwqrfnshgt5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	LastName	last_name	SingleLineText	character varying	\N	\N	20	2	f	t	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.922419+02	2022-05-18 11:24:11.922419+02
cl_6xpxnl95np7q1z	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	BaseId	base_id	SingleLineText	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.942447+02	2022-05-18 11:24:11.942447+02
cl_ijd8a2zvsvzxdb	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Title	title	SingleLineText	character varying	\N	\N	45	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	45	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.947919+02	2022-05-18 11:24:11.947919+02
cl_ttu9qrvvb56wps	ds_aej9ao4mzw9cff	p_41297240e6of48	md_nj2q6tfni3b5a8	Title	title	SingleLineText	character varying	\N	\N	255	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.955746+02	2022-05-18 11:24:11.955746+02
cl_onax81ltzbvj6l	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	Meta	meta	LongText	text	\N	\N	\N	3	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.055653+02	2022-05-18 11:24:12.055653+02
cl_kmwlqydz7qisje	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	ProjectId	project_id	ForeignKey	character varying	\N	\N	128	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	t	3	2022-05-18 11:24:12.058763+02	2022-05-18 11:24:12.058763+02
cl_vnjldzdp0ouqfw	ds_aej9ao4mzw9cff	p_41297240e6of48	md_s9jny9njrhdnwz	UnitPrice	unit_price	Decimal	real	24	\N	\N	3	f	t	t	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.075863+02	2022-05-18 11:24:12.075863+02
cl_4t4w33da2a7h6i	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	Title	title	SingleLineText	character varying	\N	\N	30	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	30	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.114015+02	2022-05-18 11:24:12.114015+02
cl_p0r4ptz7y81v5n	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5hjnvbjhgfiz6r	FkLookupColumnId	fk_lookup_column_id	ForeignKey	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	4	2022-05-18 11:24:12.117698+02	2022-05-18 11:24:12.117698+02
cl_6a13znk1jszq23	ds_aej9ao4mzw9cff	p_41297240e6of48	md_zmhiordb5vpwkf	Uuid	uuid	SingleLineText	character varying	\N	\N	255	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.126529+02	2022-05-18 11:24:12.126529+02
cl_enjcwnqd9lm0hl	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wqy1a8c5ggpyea	FkViewId	fk_view_id	ForeignKey	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	4	2022-05-18 11:24:12.140001+02	2022-05-18 11:24:12.140001+02
cl_9j9i6wt5dojluu	ds_aej9ao4mzw9cff	p_41297240e6of48	md_91fcdvmtveuz77	Role	role	SingleLineText	character varying	\N	\N	45	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	45	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.20676+02	2022-05-18 11:24:12.20676+02
cl_eb5dfr8wdn1kne	ds_aej9ao4mzw9cff	p_41297240e6of48	md_tm1aur9f3yn2l2	Deleted	deleted	Checkbox	boolean	\N	\N	\N	6	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.279397+02	2022-05-18 11:24:12.279397+02
cl_4xfv5oniyd1uke	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Version	version	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.287136+02	2022-05-18 11:24:12.287136+02
cl_elxzenlpnoman8	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	Lastname	lastname	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.302693+02	2022-05-18 11:24:12.302693+02
cl_imu0a0bfg69jjz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	HireDate	hire_date	Date	date	\N	\N	\N	7	f	\N	f	f	\N	f	\N	\N	\N	\N	date	0	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.33923+02	2022-05-18 11:24:12.33923+02
cl_st7iuxws65atp5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	FkRelatedModelId	fk_related_model_id	ForeignKey	character varying	\N	\N	20	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	7	2022-05-18 11:24:12.345788+02	2022-05-18 11:24:12.345788+02
cl_1ddmdm9b8op8jw	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	Help	help	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.418495+02	2022-05-18 11:24:12.418495+02
cl_oxsn7xt4obc7r2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	LogicalOp	logical_op	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.422165+02	2022-05-18 11:24:12.422165+02
cl_dri9nl4xksb0u6	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	FkChildColumnId	fk_child_column_id	ForeignKey	character varying	\N	\N	20	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	8	2022-05-18 11:24:12.425016+02	2022-05-18 11:24:12.425016+02
cl_lxhi365f4uspl0	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Status	status	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	'install'::character varying	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.430412+02	2022-05-18 11:24:12.430412+02
cl_ybvnp05fszzbqg	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	Order	order	Decimal	real	24	\N	\N	8	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.434409+02	2022-05-18 11:24:12.434409+02
cl_0tsvdxn69tvyh0	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	Schema	schema	LongText	text	\N	\N	\N	8	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.442311+02	2022-05-18 11:24:12.442311+02
cl_mz44l54qkdvw8m	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	Description	description	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.477272+02	2022-05-18 11:24:12.477272+02
cl_dp5y1a7pcz0s54	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	FkParentColumnId	fk_parent_column_id	ForeignKey	character varying	\N	\N	20	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	9	2022-05-18 11:24:12.484668+02	2022-05-18 11:24:12.484668+02
cl_2hvbzdhayp3m4d	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	ShipName	ship_name	SingleLineText	character varying	\N	\N	40	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	40	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.504765+02	2022-05-18 11:24:12.504765+02
cl_rw8jc866mp0w5o	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	Country	country	SingleLineText	character varying	\N	\N	15	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	15	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.507421+02	2022-05-18 11:24:12.507421+02
cl_8bww0dmnwaktnp	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	ShowBlankForm	show_blank_form	Checkbox	boolean	\N	\N	\N	11	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.581397+02	2022-05-18 11:24:12.581397+02
cl_qkxcxycznt8kxs	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	12	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.635175+02	2022-05-18 11:24:12.635175+02
cl_hkd55j8p8gdukv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	12	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.638256+02	2022-05-18 11:24:12.638256+02
cl_y2q5dfwbxx1n58	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	Show	show	Checkbox	boolean	\N	\N	\N	12	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.645888+02	2022-05-18 11:24:12.645888+02
cl_q1v7cutyjfyvov	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	13	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.664573+02	2022-05-18 11:24:12.664573+02
cl_o6znciqwr4iuto	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Rqd	rqd	Checkbox	boolean	\N	\N	\N	15	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	15	2022-05-18 11:24:12.71923+02	2022-05-18 11:24:12.71923+02
cl_g9g36fknoddzj2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	TriggeredBy	triggered_by	SingleLineText	character varying	\N	\N	255	17	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	17	2022-05-18 11:24:12.756562+02	2022-05-18 11:24:12.756562+02
cl_5em2fkdhoklrs5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	18	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	18	2022-05-18 11:24:12.770674+02	2022-05-18 11:24:12.770674+02
cl_yk84aiwkkjj5fq	ds_aej9ao4mzw9cff	p_41297240e6of48	md_79w8haiknxbkzd	CategoryName	category_name	SingleLineText	character varying	\N	\N	15	2	f	t	t	f	\N	f	\N	\N	\N	\N	character varying	15	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.886685+02	2022-05-18 11:24:11.886685+02
cl_waudmda08mbill	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	RefDbAlias	ref_db_alias	SingleLineText	character varying	\N	\N	255	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.92565+02	2022-05-18 11:24:11.92565+02
cl_bwrx80dhb9qhjd	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	BaseId	base_id	SingleLineText	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.938403+02	2022-05-18 11:24:11.938403+02
cl_obsjzit9bzkszx	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	BaseId	base_id	SingleLineText	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.94719+02	2022-05-18 11:24:11.94719+02
cl_5l0iqbng3qv1wo	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	FkUserId	fk_user_id	ForeignKey	character varying	\N	\N	20	2	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	2	2022-05-18 11:24:11.951326+02	2022-05-18 11:24:11.951326+02
cl_zodcdlchikjfqz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	Title	title	SingleLineText	character varying	\N	\N	255	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.954825+02	2022-05-18 11:24:11.954825+02
cl_0ae4rqt28h623c	ds_aej9ao4mzw9cff	p_41297240e6of48	md_c1y2qwn3okt75w	IsLocked	is_locked	Number	integer	32	0	\N	2	f	t	f	f	\N	f	\N	\N	\N	\N	integer	32	0	f	\N	\N	\N	f	2	2022-05-18 11:24:11.994146+02	2022-05-18 11:24:11.994146+02
cl_x9rp3wek2sqmkr	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.021043+02	2022-05-18 11:24:12.021043+02
cl_xdp0ttgvijpllq	ds_aej9ao4mzw9cff	p_41297240e6of48	md_tm1aur9f3yn2l2	Formula	formula	LongText	text	\N	\N	\N	3	f	\N	t	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.024829+02	2022-05-18 11:24:12.024829+02
cl_hvpo5wfg6q9bhy	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.041832+02	2022-05-18 11:24:12.041832+02
cl_3eagrih1e0zl3n	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wjrbh5du4ifq5b	Phone	phone	SingleLineText	character varying	\N	\N	24	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	24	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.08099+02	2022-05-18 11:24:12.08099+02
cl_77iddmy4xahfwa	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	BaseId	base_id	SingleLineText	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.125587+02	2022-05-18 11:24:12.125587+02
cl_e0qvng7dplmklv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	NextEnabled	next_enabled	Checkbox	boolean	\N	\N	\N	4	f	t	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.135816+02	2022-05-18 11:24:12.135816+02
cl_nce63jcfgdcfn3	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Title	title	SingleLineText	character varying	\N	\N	255	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.19735+02	2022-05-18 11:24:12.19735+02
cl_a2e7sflk8yq825	ds_aej9ao4mzw9cff	p_41297240e6of48	md_rmg6p4gh772ip8	RollupFunction	rollup_function	SingleLineText	character varying	\N	\N	255	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.20064+02	2022-05-18 11:24:12.20064+02
cl_8rjar6mb3mbmne	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	Title	title	SingleLineText	character varying	\N	\N	255	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.224094+02	2022-05-18 11:24:12.224094+02
cl_3sv6p41h8dk62h	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	Firstname	firstname	SingleLineText	character varying	\N	\N	255	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.232444+02	2022-05-18 11:24:12.232444+02
cl_zgif7qh1drm329	ds_aej9ao4mzw9cff	p_41297240e6of48	md_s9jny9njrhdnwz	Discount	discount	Decimal	real	24	\N	\N	5	f	\N	t	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.242688+02	2022-05-18 11:24:12.242688+02
cl_wnu5t8ebtmy6o5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	ColumnName	column_name	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.271012+02	2022-05-18 11:24:12.271012+02
cl_kuayl8wxyb2k5d	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	Uuid	uuid	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.27745+02	2022-05-18 11:24:12.27745+02
cl_b8tbs19b6ei4oe	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	Event	event	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.286444+02	2022-05-18 11:24:12.286444+02
cl_b45jar9oybu3vc	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	Uuid	uuid	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.290097+02	2022-05-18 11:24:12.290097+02
cl_hgpa783mbqeniv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_1wgrn82853mqsm	Enabled	enabled	Checkbox	boolean	\N	\N	\N	6	f	\N	f	f	\N	f	\N	true	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.299176+02	2022-05-18 11:24:12.299176+02
cl_jd7stjnewfms0c	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	Region	region	SingleLineText	character varying	\N	\N	15	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	15	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.320949+02	2022-05-18 11:24:12.320949+02
cl_0t5gqpzq3rmbub	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	Label	label	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.371128+02	2022-05-18 11:24:12.371128+02
cl_ysuzqfhvklkqw2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	OpType	op_type	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.413699+02	2022-05-18 11:24:12.413699+02
cl_pukj4zblib5hya	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	Public	public	Checkbox	boolean	\N	\N	\N	8	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.424204+02	2022-05-18 11:24:12.424204+02
cl_fvanz3lcymlorp	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	Country	country	SingleLineText	character varying	\N	\N	15	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	15	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.459471+02	2022-05-18 11:24:12.459471+02
cl_t3yqx8owhtr4p7	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	OpSubType	op_sub_type	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.476271+02	2022-05-18 11:24:12.476271+02
cl_38ocyrkyz7o69r	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	StatusDetails	status_details	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.489737+02	2022-05-18 11:24:12.489737+02
cl_xbrmwc3imzej7e	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	ReorderLevel	reorder_level	Number	smallint	16	0	\N	9	f	\N	f	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	f	9	2022-05-18 11:24:12.503947+02	2022-05-18 11:24:12.503947+02
cl_o5umya2z9dueqy	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	10	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.535777+02	2022-05-18 11:24:12.535777+02
cl_kki4i38mfv3zw4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	Order	order	Decimal	real	24	\N	\N	10	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.542111+02	2022-05-18 11:24:12.542111+02
cl_amisrr6l4xmpft	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	Roles	roles	SingleLineText	character varying	\N	\N	255	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.548372+02	2022-05-18 11:24:12.548372+02
cl_k7kqyx1mo2y94k	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	Discontinued	discontinued	Number	integer	32	0	\N	10	f	\N	t	f	\N	f	\N	\N	\N	\N	integer	32	0	f	\N	\N	\N	f	10	2022-05-18 11:24:12.554484+02	2022-05-18 11:24:12.554484+02
cl_j8lzgbvoagbw4d	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Clen	clen	SingleLineText	character varying	\N	\N	255	11	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.580654+02	2022-05-18 11:24:12.580654+02
cl_llodox6wyofvr5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	11	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.586493+02	2022-05-18 11:24:12.586493+02
cl_k2zpyuzl42o0ec	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	ShipRegion	ship_region	SingleLineText	character varying	\N	\N	15	12	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	15	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.648984+02	2022-05-18 11:24:12.648984+02
cl_4p1x15zwndeksm	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	Ur	ur	SingleLineText	character varying	\N	\N	255	13	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.674549+02	2022-05-18 11:24:12.674549+02
cl_viup64dwqpa7qw	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	CompanyName	company_name	SingleLineText	character varying	\N	\N	40	2	f	t	t	f	\N	f	\N	\N	\N	\N	character varying	40	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.89969+02	2022-05-18 11:24:11.89969+02
cl_r39qpsifh4dd16	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	FkViewId	fk_view_id	ForeignKey	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	2	2022-05-18 11:24:11.937008+02	2022-05-18 11:24:11.937008+02
cl_1jvx4978ntayw5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	2	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.944488+02	2022-05-18 11:24:11.944488+02
cl_37jdt2wdmn2fsh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	BaseId	base_id	SingleLineText	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.94917+02	2022-05-18 11:24:11.94917+02
cl_ygmn763v7bzofp	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	FkViewId	fk_view_id	ForeignKey	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	2	2022-05-18 11:24:11.95721+02	2022-05-18 11:24:11.95721+02
cl_o1e6spelqpafwg	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	BaseId	base_id	ForeignKey	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	2	2022-05-18 11:24:11.961029+02	2022-05-18 11:24:11.961029+02
cl_1msk1km2tcqyww	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	BaseId	base_id	SingleLineText	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.983365+02	2022-05-18 11:24:11.983365+02
cl_cp91gcskjlxt5k	ds_aej9ao4mzw9cff	p_41297240e6of48	md_89v98o1rpi4idv	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	3	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.056676+02	2022-05-18 11:24:12.056676+02
cl_f2d7hc4bji50xx	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wk0cg5v1lnk9ew	StateAbbr	state_abbr	SingleLineText	character varying	\N	\N	2	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	2	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.073786+02	2022-05-18 11:24:12.073786+02
cl_fifa1pn2w5hhiv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_j409llgiqo90yz	RegionId	region_id	ForeignKey	smallint	16	0	\N	3	f	\N	t	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	t	3	2022-05-18 11:24:12.083167+02	2022-05-18 11:24:12.083167+02
cl_nv2qb0bqfi14xj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	BaseId	base_id	ForeignKey	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	4	2022-05-18 11:24:12.11864+02	2022-05-18 11:24:12.11864+02
cl_3gm2h79uxhdzcc	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	FkViewId	fk_view_id	ForeignKey	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	4	2022-05-18 11:24:12.121371+02	2022-05-18 11:24:12.121371+02
cl_0cp6rgcqobi280	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	FkViewId	fk_view_id	ForeignKey	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	4	2022-05-18 11:24:12.138035+02	2022-05-18 11:24:12.138035+02
cl_w10l8tx9zyh2wb	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	QueryParams	query_params	LongText	text	\N	\N	\N	4	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.14109+02	2022-05-18 11:24:12.14109+02
cl_mtad5g6leynhzh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wk0cg5v1lnk9ew	StateRegion	state_region	SingleLineText	character varying	\N	\N	50	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	50	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.163626+02	2022-05-18 11:24:12.163626+02
cl_mae86z5veq16zm	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	Address	address	SingleLineText	character varying	\N	\N	60	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	60	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.180031+02	2022-05-18 11:24:12.180031+02
cl_8y5bfy3i2yupmi	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Rating	rating	Decimal	real	24	\N	\N	5	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.211629+02	2022-05-18 11:24:12.211629+02
cl_2l9fbdjfosuadr	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	FkColumnId	fk_column_id	ForeignKey	character varying	\N	\N	20	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	5	2022-05-18 11:24:12.216084+02	2022-05-18 11:24:12.216084+02
cl_jg8un0z5v4mbq0	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	FkColumnId	fk_column_id	ForeignKey	character varying	\N	\N	20	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	5	2022-05-18 11:24:12.219541+02	2022-05-18 11:24:12.219541+02
cl_6knh4wr0379knt	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	Description	description	LongText	text	\N	\N	\N	5	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.222568+02	2022-05-18 11:24:12.222568+02
cl_9g73625z1ka2js	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	Address	address	SingleLineText	character varying	\N	\N	60	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	60	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.245947+02	2022-05-18 11:24:12.245947+02
cl_blr2fs7juvjvln	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	CoverImageIdx	cover_image_idx	Number	integer	32	0	\N	6	f	\N	f	f	\N	f	\N	\N	\N	\N	integer	32	0	f	\N	\N	\N	f	6	2022-05-18 11:24:12.284906+02	2022-05-18 11:24:12.284906+02
cl_0gxxah1hq9herr	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wqy1a8c5ggpyea	Direction	direction	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	'false'::character varying	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.288676+02	2022-05-18 11:24:12.288676+02
cl_iszr4mbjyqie02	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	City	city	SingleLineText	character varying	\N	\N	15	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	15	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.317177+02	2022-05-18 11:24:12.317177+02
cl_atylefeb8f6yfh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_rmg6p4gh772ip8	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	7	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.341267+02	2022-05-18 11:24:12.341267+02
cl_6c3qkavaz14cfj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_91fcdvmtveuz77	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	7	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.347984+02	2022-05-18 11:24:12.347984+02
cl_l4uk1f2f7g5wxq	ds_aej9ao4mzw9cff	p_41297240e6of48	md_tm1aur9f3yn2l2	Order	order	Decimal	real	24	\N	\N	7	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.352599+02	2022-05-18 11:24:12.352599+02
cl_f3lqniknfi47bd	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	UnitsInStock	units_in_stock	Number	smallint	16	0	\N	7	f	\N	f	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	f	7	2022-05-18 11:24:12.387804+02	2022-05-18 11:24:12.387804+02
cl_8ux969grq7yqo6	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	PostalCode	postal_code	SingleLineText	character varying	\N	\N	10	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	10	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.398254+02	2022-05-18 11:24:12.398254+02
cl_r0os9n1us2h7be	ds_aej9ao4mzw9cff	p_41297240e6of48	md_rmg6p4gh772ip8	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	8	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.416035+02	2022-05-18 11:24:12.416035+02
cl_3zglox2j8sjgrz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_91fcdvmtveuz77	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	8	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.419907+02	2022-05-18 11:24:12.419907+02
cl_epm4ujkyxsksgv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	InflectionColumn	inflection_column	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.423187+02	2022-05-18 11:24:12.423187+02
cl_sqj0wjnlu6z8j8	ds_aej9ao4mzw9cff	p_41297240e6of48	md_tm1aur9f3yn2l2	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	8	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.426823+02	2022-05-18 11:24:12.426823+02
cl_hn4ew82cx6hg2s	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wqy1a8c5ggpyea	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	8	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.43192+02	2022-05-18 11:24:12.43192+02
cl_5qy65ad5mlzglb	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Type	type	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.436181+02	2022-05-18 11:24:12.436181+02
cl_grx4o67kbsjtct	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	PostalCode	postal_code	SingleLineText	character varying	\N	\N	10	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	10	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.455006+02	2022-05-18 11:24:12.455006+02
cl_qy73tlix63j4b7	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Np	np	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.480013+02	2022-05-18 11:24:12.480013+02
cl_bmdab93tjhr67q	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wqy1a8c5ggpyea	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	9	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.488417+02	2022-05-18 11:24:12.488417+02
cl_ds7fd5osbe41sp	ds_aej9ao4mzw9cff	p_41297240e6of48	md_91fcdvmtveuz77	BaseId	base_id	SingleLineText	character varying	\N	\N	20	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.928144+02	2022-05-18 11:24:11.928144+02
cl_2vsljvxzziveo3	ds_aej9ao4mzw9cff	p_41297240e6of48	md_0z5ofu8o9xph1j	Title	title	SingleLineText	character varying	\N	\N	255	2	f	t	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.956504+02	2022-05-18 11:24:11.956504+02
cl_x7rpf6qkm426p4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	CompanyName	company_name	SingleLineText	character varying	\N	\N	40	2	f	t	t	f	\N	f	\N	\N	\N	\N	character varying	40	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.979429+02	2022-05-18 11:24:11.979429+02
cl_cxqf2kqwmat0k3	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5hjnvbjhgfiz6r	FkRelationColumnId	fk_relation_column_id	ForeignKey	character varying	\N	\N	20	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	3	2022-05-18 11:24:12.022899+02	2022-05-18 11:24:12.022899+02
cl_fw3v3fz91x4v6e	ds_aej9ao4mzw9cff	p_41297240e6of48	md_91fcdvmtveuz77	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.033236+02	2022-05-18 11:24:12.033236+02
cl_yldbutsud8s8es	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	ProjectId	project_id	SingleLineText	character varying	\N	\N	128	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	128	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.049317+02	2022-05-18 11:24:12.049317+02
cl_hrs9ijvoaus9ls	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	Password	password	SingleLineText	character varying	\N	\N	255	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.064787+02	2022-05-18 11:24:12.064787+02
cl_m6bwhgghyxz32g	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	FkHookId	fk_hook_id	SingleLineText	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.128181+02	2022-05-18 11:24:12.128181+02
cl_2etavxey6fg0fo	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	TableName	table_name	SingleLineText	character varying	\N	\N	255	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.145518+02	2022-05-18 11:24:12.145518+02
cl_c0u4gz4xsq5zqu	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	Subheading	subheading	SingleLineText	character varying	\N	\N	255	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.202148+02	2022-05-18 11:24:12.202148+02
cl_b86m0u9nsa08p9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5hjnvbjhgfiz6r	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	6	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.272299+02	2022-05-18 11:24:12.272299+02
cl_hqmb5f2hlm8kbt	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	Uuid	uuid	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.280232+02	2022-05-18 11:24:12.280232+02
cl_b4f6kuyhpgnxf6	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	Type	type	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	'table'::character varying	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.296766+02	2022-05-18 11:24:12.296766+02
cl_uhtznd7f8pokfx	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	Label	label	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.347224+02	2022-05-18 11:24:12.347224+02
cl_xx7o3hptppb3jb	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	Type	type	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.351455+02	2022-05-18 11:24:12.351455+02
cl_5piiaqr116azp8	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	Operation	operation	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.356957+02	2022-05-18 11:24:12.356957+02
cl_jcmfany6r3buss	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	Color	color	SingleLineText	character varying	\N	\N	255	7	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.372708+02	2022-05-18 11:24:12.372708+02
cl_8cf2j5ptfccdk1	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	Password	password	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.435666+02	2022-05-18 11:24:12.435666+02
cl_uwmjhs1h96b5j7	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	ShowSystemFields	show_system_fields	Checkbox	boolean	\N	\N	\N	8	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.454455+02	2022-05-18 11:24:12.454455+02
cl_0mczoi3l3e86wz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	Payload	payload	Checkbox	boolean	\N	\N	\N	9	f	\N	f	f	\N	f	\N	true	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.485383+02	2022-05-18 11:24:12.485383+02
cl_tzcg2k5pfnkv2c	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Event	event	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.491223+02	2022-05-18 11:24:12.491223+02
cl_olwudcee6vaydm	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	OpenedDate	opened_date	DateTime	timestamp with time zone	\N	\N	\N	10	f	\N	f	f	\N	f	\N	\N	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.545431+02	2022-05-18 11:24:12.545431+02
cl_nj629z71fvb8xr	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	RestrictNumber	restrict_number	SingleLineText	character varying	\N	\N	255	11	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.58493+02	2022-05-18 11:24:12.58493+02
cl_myyzrukjep7quc	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	ResetPasswordExpires	reset_password_expires	DateTime	timestamp with time zone	\N	\N	\N	11	f	\N	f	f	\N	f	\N	\N	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.59899+02	2022-05-18 11:24:12.59899+02
cl_87e0ickdzoj44q	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	FkMmParentColumnId	fk_mm_parent_column_id	ForeignKey	character varying	\N	\N	20	12	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	12	2022-05-18 11:24:12.635953+02	2022-05-18 11:24:12.635953+02
cl_bsjstv693k5gdj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	ShipPostalCode	ship_postal_code	SingleLineText	character varying	\N	\N	10	13	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	10	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.684124+02	2022-05-18 11:24:12.684124+02
cl_nacdeb0qxxghy1	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	14	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.696845+02	2022-05-18 11:24:12.696845+02
cl_xty7soa1r3z01z	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	Dr	dr	SingleLineText	character varying	\N	\N	255	14	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.702211+02	2022-05-18 11:24:12.702211+02
cl_agwmleh0fxpuyj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	15	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	15	2022-05-18 11:24:12.730234+02	2022-05-18 11:24:12.730234+02
cl_n7bvxl6hrnzucd	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	16	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	16	2022-05-18 11:24:12.750022+02	2022-05-18 11:24:12.750022+02
cl_wrshd8cmtqyzk5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_89v98o1rpi4idv	NcOrgsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	5	2022-05-18 11:24:12.865332+02	2022-05-18 11:24:12.865332+02
cl_8cezepfnbj7t2k	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wqy1a8c5ggpyea	NcColumnsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	10	2022-05-18 11:24:12.870321+02	2022-05-18 11:24:12.870321+02
cl_6skr9pfy1fpbdh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	ProductsList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	13	2022-05-18 11:24:12.872968+02	2022-05-18 11:24:12.872968+02
cl_vr6w15ow265mdm	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	NcProjectUsersV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	18	2022-05-18 11:24:12.875724+02	2022-05-18 11:24:12.875724+02
cl_hzx263qp16kmx1	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	NcProjectsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	14	2022-05-18 11:24:12.975568+02	2022-05-18 11:24:12.975568+02
cl_l0q6rlkn5vn5ig	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wqy1a8c5ggpyea	NcViewsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	11	2022-05-18 11:24:12.98247+02	2022-05-18 11:24:12.98247+02
cl_lm7p4u967pdvrv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	NcColRelationsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	18	2022-05-18 11:24:12.986613+02	2022-05-18 11:24:12.986613+02
cl_f3nxqy328zxk7v	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	NcFormViewV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	18	2022-05-18 11:24:13.060909+02	2022-05-18 11:24:13.060909+02
cl_fpxyzzfs6zlpdw	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	NcGalleryViewV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	19	2022-05-18 11:24:13.098707+02	2022-05-18 11:24:13.098707+02
cl_x801ithwdkcrx9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	ProductName	product_name	SingleLineText	character varying	\N	\N	40	2	f	t	t	f	\N	f	\N	\N	\N	\N	character varying	40	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.974353+02	2022-05-18 11:24:11.974353+02
cl_bbui77uct3fuc5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_j409llgiqo90yz	TerritoryDescription	territory_description	SingleLineText	character	\N	\N	\N	2	f	t	t	f	\N	f	\N	\N	\N	\N	character	\N	\N	f	\N	\N	\N	f	2	2022-05-18 11:24:11.982877+02	2022-05-18 11:24:11.982877+02
cl_dan2jfbs2nmq5u	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	FirstName	first_name	SingleLineText	character varying	\N	\N	10	3	f	\N	t	f	\N	f	\N	\N	\N	\N	character varying	10	\N	f	\N	\N	\N	f	3	2022-05-18 11:24:12.02261+02	2022-05-18 11:24:12.02261+02
cl_2jbc6l8uccbcxv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	FkColumnId	fk_column_id	ForeignKey	character varying	\N	\N	20	3	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	3	2022-05-18 11:24:12.036739+02	2022-05-18 11:24:12.036739+02
cl_v2ll0vs8jz4501	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	FkViewId	fk_view_id	ForeignKey	character varying	\N	\N	20	3	t	\N	t	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	3	2022-05-18 11:24:12.043945+02	2022-05-18 11:24:12.043945+02
cl_2n98v3z6z9y6eo	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	EmployeeId	employee_id	ForeignKey	smallint	16	0	\N	3	f	\N	f	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	t	3	2022-05-18 11:24:12.068247+02	2022-05-18 11:24:12.068247+02
cl_w9k4oudkzve2tu	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	FkModelId	fk_model_id	ForeignKey	character varying	\N	\N	20	4	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	4	2022-05-18 11:24:12.137333+02	2022-05-18 11:24:12.137333+02
cl_lst1bydbilrj0f	ds_aej9ao4mzw9cff	p_41297240e6of48	md_0z5ofu8o9xph1j	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	4	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	4	2022-05-18 11:24:12.140777+02	2022-05-18 11:24:12.140777+02
cl_es4sn3km17zc4e	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5hjnvbjhgfiz6r	Deleted	deleted	Checkbox	boolean	\N	\N	\N	5	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.198821+02	2022-05-18 11:24:12.198821+02
cl_to8frc2rpu5ar7	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	Type	type	SingleLineText	character varying	\N	\N	255	5	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	5	2022-05-18 11:24:12.210497+02	2022-05-18 11:24:12.210497+02
cl_t9bju32ijpe5ei	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	BirthDate	birth_date	Date	date	\N	\N	\N	6	f	\N	f	f	\N	f	\N	\N	\N	\N	date	0	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.269277+02	2022-05-18 11:24:12.269277+02
cl_6fmcncpfbx7wy0	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	SuccessMsg	success_msg	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.275311+02	2022-05-18 11:24:12.275311+02
cl_m2jfcpnqk7qb4p	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	Group	group	SingleLineText	character varying	\N	\N	255	6	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	6	2022-05-18 11:24:12.291477+02	2022-05-18 11:24:12.291477+02
cl_adhc5aoe5s2dkl	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	Meta	meta	LongText	text	\N	\N	\N	7	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	7	2022-05-18 11:24:12.376841+02	2022-05-18 11:24:12.376841+02
cl_bi5wfcj36qnihx	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	TestCall	test_call	Checkbox	boolean	\N	\N	\N	8	f	\N	f	f	\N	f	\N	true	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.425826+02	2022-05-18 11:24:12.425826+02
cl_wo85353duh2q82	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	Freight	freight	Decimal	real	24	\N	\N	8	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.447918+02	2022-05-18 11:24:12.447918+02
cl_tox2sh660xnbo4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	InflectionTable	inflection_table	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.482357+02	2022-05-18 11:24:12.482357+02
cl_c18gv46qxxr0nd	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	Deleted	deleted	Checkbox	boolean	\N	\N	\N	9	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.495254+02	2022-05-18 11:24:12.495254+02
cl_hpktuwr8acmf6w	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	ShowAllFields	show_all_fields	Checkbox	boolean	\N	\N	\N	10	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.534476+02	2022-05-18 11:24:12.534476+02
cl_f4nw9fnghi0wi2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	Description	description	LongText	text	\N	\N	\N	11	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.582099+02	2022-05-18 11:24:12.582099+02
cl_rzvj3urxhsfmli	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	FkMmChildColumnId	fk_mm_child_column_id	ForeignKey	character varying	\N	\N	20	11	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	20	\N	f	\N	\N	\N	t	11	2022-05-18 11:24:12.587655+02	2022-05-18 11:24:12.587655+02
cl_eo23q73pocm3d5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	11	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.594331+02	2022-05-18 11:24:12.594331+02
cl_p83aw566opsxqw	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	Public	public	Checkbox	boolean	\N	\N	\N	12	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.632514+02	2022-05-18 11:24:12.632514+02
cl_tespjgc8eqnn3u	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	Pinned	pinned	Checkbox	boolean	\N	\N	\N	12	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.638953+02	2022-05-18 11:24:12.638953+02
cl_eagzbs90wvmutu	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	14	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.691994+02	2022-05-18 11:24:12.691994+02
cl_3phwdvhn8yfixc	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	15	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	15	2022-05-18 11:24:12.730623+02	2022-05-18 11:24:12.730623+02
cl_s4i6nmg47h2j5p	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	Notes	notes	LongText	text	\N	\N	\N	16	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	16	2022-05-18 11:24:12.740447+02	2022-05-18 11:24:12.740447+02
cl_ppkvlb1bhb29sv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	16	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	16	2022-05-18 11:24:12.752355+02	2022-05-18 11:24:12.752355+02
cl_a7kx6brjog1as9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_79w8haiknxbkzd	ProductsList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	5	2022-05-18 11:24:12.86515+02	2022-05-18 11:24:12.86515+02
cl_ltn26rjondvorz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_tm1aur9f3yn2l2	NcColumnsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	10	2022-05-18 11:24:12.870078+02	2022-05-18 11:24:12.870078+02
cl_3rcpv6o301oikz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	NcViewsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	13	2022-05-18 11:24:12.872725+02	2022-05-18 11:24:12.872725+02
cl_j32rfmzk6jy8en	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	NcFormViewColumnsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	17	2022-05-18 11:24:12.875421+02	2022-05-18 11:24:12.875421+02
cl_yuvf9ho2j6gyx2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_q2z56jvps22ukx	TerritoriesRead	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	4	2022-05-18 11:24:12.969162+02	2022-05-18 11:24:12.969162+02
cl_xm8dld9std1g53	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	NcGalleryViewV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	14	2022-05-18 11:24:12.975145+02	2022-05-18 11:24:12.975145+02
cl_8jwbbixyl9l1gu	ds_aej9ao4mzw9cff	p_41297240e6of48	md_89v98o1rpi4idv	NcUsersV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	6	2022-05-18 11:24:12.982107+02	2022-05-18 11:24:12.982107+02
cl_6nefo9nk04gtus	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	NcColumnsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	16	2022-05-18 11:24:12.986321+02	2022-05-18 11:24:12.986321+02
cl_iih5ukipx3n3mn	ds_aej9ao4mzw9cff	p_41297240e6of48	md_rmg6p4gh772ip8	NcColumnsV2Read2	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	11	2022-05-18 11:24:13.053024+02	2022-05-18 11:24:13.053024+02
cl_9zapplyeoqzc3w	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	OrdersList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	21	2022-05-18 11:24:13.060103+02	2022-05-18 11:24:13.060103+02
cl_ru32bw7buii6la	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	EmployeesRead	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	22	2022-05-18 11:24:13.09779+02	2022-05-18 11:24:13.09779+02
cl_jny174p7hh10qk	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	Help	help	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.437429+02	2022-05-18 11:24:12.437429+02
cl_mcx1wq0b8d1q8n	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	RefreshToken	refresh_token	SingleLineText	character varying	\N	\N	255	8	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	8	2022-05-18 11:24:12.440792+02	2022-05-18 11:24:12.440792+02
cl_i2pwveeqruhjtz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	Password	password	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.483566+02	2022-05-18 11:24:12.483566+02
cl_dpp1unrz3q1c3g	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	Enabled	enabled	Checkbox	boolean	\N	\N	\N	9	f	\N	f	f	\N	f	\N	true	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.500921+02	2022-05-18 11:24:12.500921+02
cl_v7v1onya6tagss	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	Show	show	Checkbox	boolean	\N	\N	\N	10	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.5329+02	2022-05-18 11:24:12.5329+02
cl_sffsxt353w499o	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	Conditions	conditions	LongText	text	\N	\N	\N	10	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.538964+02	2022-05-18 11:24:12.538964+02
cl_qiubigzgwgau18	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	Mm	mm	Checkbox	boolean	\N	\N	\N	10	f	\N	f	f	\N	f	\N	false	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.551452+02	2022-05-18 11:24:12.551452+02
cl_rqltw519nd1bvk	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	Uuid	uuid	SingleLineText	character varying	\N	\N	255	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.558308+02	2022-05-18 11:24:12.558308+02
cl_n2appwjzrymdk5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	ShipAddress	ship_address	SingleLineText	character varying	\N	\N	60	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	60	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.562729+02	2022-05-18 11:24:12.562729+02
cl_qsksqgb6y745zn	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	11	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.588874+02	2022-05-18 11:24:12.588874+02
cl_2nyio17f5epmq0	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	Deleted	deleted	Checkbox	boolean	\N	\N	\N	11	f	\N	f	f	\N	f	\N	false	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.591958+02	2022-05-18 11:24:12.591958+02
cl_oorghne0ybyrmk	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	Country	country	SingleLineText	character varying	\N	\N	15	12	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	15	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.620786+02	2022-05-18 11:24:12.620786+02
cl_cne8f9kkhhmz6d	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	Order	order	Decimal	real	24	\N	\N	12	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.623659+02	2022-05-18 11:24:12.623659+02
cl_10h4tlr5w8w9l9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	ErrorCode	error_code	SingleLineText	character varying	\N	\N	255	12	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.626932+02	2022-05-18 11:24:12.626932+02
cl_74xo3c6o7uhjsa	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	12	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.634254+02	2022-05-18 11:24:12.634254+02
cl_6f5tq9drcdxwh7	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	IsMeta	is_meta	Checkbox	boolean	\N	\N	\N	12	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.637256+02	2022-05-18 11:24:12.637256+02
cl_wy8brvdnpgy3d3	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Pv	pv	Checkbox	boolean	\N	\N	\N	14	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.693983+02	2022-05-18 11:24:12.693983+02
cl_nzd6ey7esqq626	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	ShipCountry	ship_country	SingleLineText	character varying	\N	\N	15	14	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	15	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.712487+02	2022-05-18 11:24:12.712487+02
cl_w9aupuppuv6o2k	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	Photo	photo	SpecificDBType	bytea	\N	\N	\N	15	f	\N	f	f	\N	f	\N	\N	\N	\N	bytea	\N	\N	f	\N	\N	\N	f	15	2022-05-18 11:24:12.718277+02	2022-05-18 11:24:12.718277+02
cl_p2vlwzdwz6g4t3	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	15	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	15	2022-05-18 11:24:12.720876+02	2022-05-18 11:24:12.720876+02
cl_2z8e4cfd1sj65e	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	15	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	15	2022-05-18 11:24:12.726798+02	2022-05-18 11:24:12.726798+02
cl_sw2ka2n7mjx46z	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	Response	response	SingleLineText	character varying	\N	\N	255	16	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	16	2022-05-18 11:24:12.741493+02	2022-05-18 11:24:12.741493+02
cl_np4oitaanopvi0	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Creator	creator	SingleLineText	character varying	\N	\N	255	16	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	16	2022-05-18 11:24:12.747928+02	2022-05-18 11:24:12.747928+02
cl_bvt77xombx4esv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	CreatorWebsite	creator_website	SingleLineText	character varying	\N	\N	255	17	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	17	2022-05-18 11:24:12.762217+02	2022-05-18 11:24:12.762217+02
cl_2jdqr32vfn4gw4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Price	price	SingleLineText	character varying	\N	\N	255	18	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	18	2022-05-18 11:24:12.77413+02	2022-05-18 11:24:12.77413+02
cl_ddia0w2o8xb2uv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	19	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	19	2022-05-18 11:24:12.78366+02	2022-05-18 11:24:12.78366+02
cl_xh3pswy5ur5ov9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	20	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	20	2022-05-18 11:24:12.788586+02	2022-05-18 11:24:12.788586+02
cl_cz3l6lgwksaq2e	ds_aej9ao4mzw9cff	p_41297240e6of48	md_q2z56jvps22ukx	EmployeesRead	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	3	2022-05-18 11:24:12.864347+02	2022-05-18 11:24:12.864347+02
cl_d8fg3jivyrrhux	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6pjpqivc2ymh3b	NcColumnsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	8	2022-05-18 11:24:12.868933+02	2022-05-18 11:24:12.868933+02
cl_j44zcituyfu32m	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	NcKanbanViewColumnsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	13	2022-05-18 11:24:12.871533+02	2022-05-18 11:24:12.871533+02
cl_p619rqzumc5ueb	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	NcColumnsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	15	2022-05-18 11:24:12.874029+02	2022-05-18 11:24:12.874029+02
cl_mg31nh0giwezx0	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	NcColumnsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	19	2022-05-18 11:24:12.877936+02	2022-05-18 11:24:12.877936+02
cl_xdmlbpk1l1lxt6	ds_aej9ao4mzw9cff	p_41297240e6of48	md_zmhiordb5vpwkf	NcViewsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	8	2022-05-18 11:24:12.973158+02	2022-05-18 11:24:12.973158+02
cl_z02hnsgf78dgxx	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	NcKanbanViewV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	14	2022-05-18 11:24:12.981229+02	2022-05-18 11:24:12.981229+02
cl_aedmuthbh2dyrz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	NcFilterExpV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	17	2022-05-18 11:24:12.985497+02	2022-05-18 11:24:12.985497+02
cl_rrv3kdmoalscjb	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5hjnvbjhgfiz6r	NcColumnsV2Read2	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	10	2022-05-18 11:24:13.057711+02	2022-05-18 11:24:13.057711+02
cl_18yvnok8ybldha	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColLookupV2List1	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	36	2022-05-18 11:24:13.065364+02	2022-05-18 11:24:13.065364+02
cl_ly7t4ezyypj5ld	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	NcFilterExpV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	18	2022-05-18 11:24:13.093895+02	2022-05-18 11:24:13.093895+02
cl_e2mhvq3w5in28m	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	NcViewsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	19	2022-05-18 11:24:13.119077+02	2022-05-18 11:24:13.119077+02
cl_hi074lxlm7gyan	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	NcTeamUsersV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	19	2022-05-18 11:24:12.988022+02	2022-05-18 11:24:13.310179+02
cl_dd7bou38xmdibz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	RestrictTypes	restrict_types	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.484347+02	2022-05-18 11:24:12.484347+02
cl_t4lk8w5lb5xsa6	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	Hidden	hidden	Decimal	real	24	\N	\N	9	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.490197+02	2022-05-18 11:24:12.490197+02
cl_x7ssnak28x8548	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	LockType	lock_type	SingleLineText	character varying	\N	\N	255	9	f	\N	f	f	\N	f	\N	'collaborative'::character varying	\N	\N	character varying	255	\N	f	\N	\N	\N	f	9	2022-05-18 11:24:12.507136+02	2022-05-18 11:24:12.507136+02
cl_qcdyzyflhbcxam	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Ns	ns	SingleLineText	character varying	\N	\N	255	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.529944+02	2022-05-18 11:24:12.529944+02
cl_ja4j53raurxvz1	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	Order	order	Decimal	real	24	\N	\N	10	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.542897+02	2022-05-18 11:24:12.542897+02
cl_1821goipxujpwj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	Fax	fax	SingleLineText	character varying	\N	\N	24	11	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	24	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.565854+02	2022-05-18 11:24:12.565854+02
cl_55duarfeg1agza	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Icon	icon	SingleLineText	character varying	\N	\N	255	11	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.589349+02	2022-05-18 11:24:12.589349+02
cl_69guec8gq6p2oe	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	12	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.621449+02	2022-05-18 11:24:12.621449+02
cl_d60otwkbbnwcl4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	Order	order	Decimal	real	24	\N	\N	12	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.624265+02	2022-05-18 11:24:12.624265+02
cl_lxa3pmvssaszo6	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Tags	tags	SingleLineText	character varying	\N	\N	255	12	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.631443+02	2022-05-18 11:24:12.631443+02
cl_ycfpzz491h9tjf	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	ResetPasswordToken	reset_password_token	SingleLineText	character varying	\N	\N	255	12	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.640749+02	2022-05-18 11:24:12.640749+02
cl_gtql7p62i2525w	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	Homepage	homepage	LongText	text	\N	\N	\N	12	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.645269+02	2022-05-18 11:24:12.645269+02
cl_md0irktw1j1uvk	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	HomePhone	home_phone	SingleLineText	character varying	\N	\N	24	13	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	24	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.664281+02	2022-05-18 11:24:12.664281+02
cl_fpfcswiqdsstff	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	Dimensions	dimensions	SingleLineText	character varying	\N	\N	255	13	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.67221+02	2022-05-18 11:24:12.67221+02
cl_rqvi328ikm9se5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	Order	order	Decimal	real	24	\N	\N	13	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.680812+02	2022-05-18 11:24:12.680812+02
cl_gvn88mg3ykq5n8	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	Order	order	Decimal	real	24	\N	\N	14	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.707034+02	2022-05-18 11:24:12.707034+02
cl_lw8tfi7qc221ob	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	FkIndexName	fk_index_name	SingleLineText	character varying	\N	\N	255	15	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	15	2022-05-18 11:24:12.732258+02	2022-05-18 11:24:12.732258+02
cl_cqgz84qtnyafgv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	Deleted	deleted	Checkbox	boolean	\N	\N	\N	16	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	16	2022-05-18 11:24:12.748772+02	2022-05-18 11:24:12.748772+02
cl_g0f1yyy9dmvuta	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	17	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	17	2022-05-18 11:24:12.763189+02	2022-05-18 11:24:12.763189+02
cl_ltaq7rwb27qjmp	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	18	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	18	2022-05-18 11:24:12.774614+02	2022-05-18 11:24:12.774614+02
cl_dyz994lxx7yesp	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wjrbh5du4ifq5b	OrdersList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	4	2022-05-18 11:24:12.864708+02	2022-05-18 11:24:12.864708+02
cl_art81q63geekql	ds_aej9ao4mzw9cff	p_41297240e6of48	md_rmg6p4gh772ip8	NcColumnsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	9	2022-05-18 11:24:12.869455+02	2022-05-18 11:24:12.869455+02
cl_xo0892py3ndpke	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	NcColumnsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	13	2022-05-18 11:24:12.872014+02	2022-05-18 11:24:12.872014+02
cl_sxpcp2cwk39cag	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	NcAuditV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	16	2022-05-18 11:24:12.874615+02	2022-05-18 11:24:12.874615+02
cl_m2p63veyy3cmww	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColFormulaV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	34	2022-05-18 11:24:12.879404+02	2022-05-18 11:24:12.879404+02
cl_63p3m5hmmo1sal	ds_aej9ao4mzw9cff	p_41297240e6of48	md_7vpsxklk3sph75	CustomersRead	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	4	2022-05-18 11:24:12.968266+02	2022-05-18 11:24:12.968266+02
cl_8bq9golhirsif4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_rmg6p4gh772ip8	NcColumnsV2Read1	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	10	2022-05-18 11:24:12.974813+02	2022-05-18 11:24:12.974813+02
cl_6dhtrez796pgcx	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5hjnvbjhgfiz6r	NcColumnsV2Read1	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	9	2022-05-18 11:24:12.981759+02	2022-05-18 11:24:12.981759+02
cl_38nvbo6hofo1ba	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	NcProjectsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	17	2022-05-18 11:24:13.059075+02	2022-05-18 11:24:13.059075+02
cl_ub0vo7g9n0zjrb	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	NcModelsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	22	2022-05-18 11:24:13.096499+02	2022-05-18 11:24:13.096499+02
cl_7s2mqpkdq9y4f5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	NcColumnsV2Read3	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	23	2022-05-18 11:24:13.120199+02	2022-05-18 11:24:13.120199+02
cl_9xl9rhlhzm8ga2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	NcColumnsV2Read4	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	24	2022-05-18 11:24:13.138596+02	2022-05-18 11:24:13.138596+02
cl_5u2tjgd4jktqj2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	NcModelsV2Read1	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	25	2022-05-18 11:24:13.150269+02	2022-05-18 11:24:13.150269+02
cl_8y5rf4iiyqjfad	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	Phone	phone	SingleLineText	character varying	\N	\N	24	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	24	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.514976+02	2022-05-18 11:24:12.514976+02
cl_xrpbnrlguot0p4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	Status	status	SingleLineText	character varying	\N	\N	255	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.531298+02	2022-05-18 11:24:12.531298+02
cl_06o2ocohphf0sv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	RestrictSize	restrict_size	SingleLineText	character varying	\N	\N	255	10	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.537787+02	2022-05-18 11:24:12.537787+02
cl_st7kgq3v75o7np	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	Order	order	Decimal	real	24	\N	\N	10	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	10	2022-05-18 11:24:12.550312+02	2022-05-18 11:24:12.550312+02
cl_9yvg1tci6tf6xg	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	PostalCode	postal_code	SingleLineText	character varying	\N	\N	10	11	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	10	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.578972+02	2022-05-18 11:24:12.578972+02
cl_al38jy5qx7kc9t	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	IsGroup	is_group	Checkbox	boolean	\N	\N	\N	11	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.58237+02	2022-05-18 11:24:12.58237+02
cl_2sucn53aey97zz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	Fax	fax	SingleLineText	character varying	\N	\N	24	11	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	24	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.604319+02	2022-05-18 11:24:12.604319+02
cl_qvavs3z81dn4d8	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	Uuid	uuid	SingleLineText	character varying	\N	\N	255	12	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.625382+02	2022-05-18 11:24:12.625382+02
cl_u1du84u11xn4q7	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	12	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.629992+02	2022-05-18 11:24:12.629992+02
cl_5x63ldetbc36fe	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Payload	payload	Checkbox	boolean	\N	\N	\N	12	f	\N	f	f	\N	f	\N	true	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.636372+02	2022-05-18 11:24:12.636372+02
cl_m1wh0xmljkxlzu	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Pk	pk	Checkbox	boolean	\N	\N	\N	13	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.662336+02	2022-05-18 11:24:12.662336+02
cl_pu8p0rjz0tzntw	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	BannerImageUrl	banner_image_url	SingleLineText	character varying	\N	\N	255	13	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.665919+02	2022-05-18 11:24:12.665919+02
cl_1ct8sqavrsm463	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	Deleted	deleted	Checkbox	boolean	\N	\N	\N	13	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.679098+02	2022-05-18 11:24:12.679098+02
cl_dvwubuqeafa6qp	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	Extension	extension	SingleLineText	character varying	\N	\N	4	14	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	4	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.69237+02	2022-05-18 11:24:12.69237+02
cl_7sopjg9jqet3lv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	14	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.697143+02	2022-05-18 11:24:12.697143+02
cl_v0swmi0cyni03y	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	EmailVerified	email_verified	Checkbox	boolean	\N	\N	\N	14	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.705665+02	2022-05-18 11:24:12.705665+02
cl_9bpgd7aysn8o4k	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Condition	condition	Checkbox	boolean	\N	\N	\N	15	f	\N	f	f	\N	f	\N	false	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	15	2022-05-18 11:24:12.725286+02	2022-05-18 11:24:12.725286+02
cl_2kj2q3s1cv02g7	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	15	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	15	2022-05-18 11:24:12.733664+02	2022-05-18 11:24:12.733664+02
cl_xkvq5go221sd4b	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	16	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	16	2022-05-18 11:24:12.746974+02	2022-05-18 11:24:12.746974+02
cl_ydq1ak4824ae8r	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	17	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	17	2022-05-18 11:24:12.760284+02	2022-05-18 11:24:12.760284+02
cl_v9cnldmlymawdy	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	OrderDetailsList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	11	2022-05-18 11:24:12.870636+02	2022-05-18 11:24:12.870636+02
cl_ftb3421h4x8xxr	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	NcGalleryViewColumnsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	17	2022-05-18 11:24:12.876039+02	2022-05-18 11:24:12.876039+02
cl_uq57oat6mrcthd	ds_aej9ao4mzw9cff	p_41297240e6of48	md_s9jny9njrhdnwz	ProductsRead	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	7	2022-05-18 11:24:12.978723+02	2022-05-18 11:24:12.978723+02
cl_ikur0598w6hz3f	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	NcModelsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	16	2022-05-18 11:24:12.983071+02	2022-05-18 11:24:12.983071+02
cl_03jw5hlwoiw2ji	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	EmployeesRead	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	17	2022-05-18 11:24:13.063386+02	2022-05-18 11:24:13.063386+02
cl_yl6pabbkd8b0yk	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	ShippersRead	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	18	2022-05-18 11:24:13.099667+02	2022-05-18 11:24:13.099667+02
cl_4gstsln64mwssn	ds_aej9ao4mzw9cff	p_41297240e6of48	md_nj2q6tfni3b5a8	NcTeamUsersV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	5	2022-05-18 11:24:12.865594+02	2022-05-18 11:24:13.308614+02
cl_h4vfb2b5flhu5c	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	Password	password	SingleLineText	character varying	\N	\N	255	11	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.604846+02	2022-05-18 11:24:12.604846+02
cl_ej7h0ko9o7l40k	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	ShipCity	ship_city	SingleLineText	character varying	\N	\N	15	11	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	15	\N	f	\N	\N	\N	f	11	2022-05-18 11:24:12.607614+02	2022-05-18 11:24:12.607614+02
cl_2pskm13svz0wnf	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	Details	details	LongText	text	\N	\N	\N	12	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	12	2022-05-18 11:24:12.625818+02	2022-05-18 11:24:12.625818+02
cl_qjegk9iz1niq0i	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	13	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.668754+02	2022-05-18 11:24:12.668754+02
cl_xlbezj3jt14886	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	LogoUrl	logo_url	SingleLineText	character varying	\N	\N	255	14	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.69738+02	2022-05-18 11:24:12.69738+02
cl_s2e5qmhhb1jten	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Headers	headers	LongText	text	\N	\N	\N	14	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.702871+02	2022-05-18 11:24:12.702871+02
cl_4y171k6kshi6e1	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	14	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.705943+02	2022-05-18 11:24:12.705943+02
cl_cou9zyhi047rfz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Input	input	LongText	text	\N	\N	\N	15	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	15	2022-05-18 11:24:12.725932+02	2022-05-18 11:24:12.725932+02
cl_tqy34owdpe210n	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	ReportsTo	reports_to	ForeignKey	smallint	16	0	\N	17	f	\N	f	f	\N	f	\N	\N	\N	\N	smallint	16	0	f	\N	\N	\N	t	17	2022-05-18 11:24:12.757345+02	2022-05-18 11:24:12.757345+02
cl_6o6s13d0gw5jk7	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	PhotoPath	photo_path	SingleLineText	character varying	\N	\N	255	18	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	18	2022-05-18 11:24:12.771876+02	2022-05-18 11:24:12.771876+02
cl_mr8rz9hd7ttjwp	ds_aej9ao4mzw9cff	p_41297240e6of48	md_91fcdvmtveuz77	NcViewsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	9	2022-05-18 11:24:12.869807+02	2022-05-18 11:24:12.869807+02
cl_2ruapfr842h3bp	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	NcUsersV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	13	2022-05-18 11:24:12.872415+02	2022-05-18 11:24:12.872415+02
cl_fzz2wpiidlwgie	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	NcDisabledModelsForRoleV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	16	2022-05-18 11:24:12.875023+02	2022-05-18 11:24:12.875023+02
cl_oe5982qxpft694	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	CategoriesRead	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	12	2022-05-18 11:24:12.976779+02	2022-05-18 11:24:12.976779+02
cl_leidqf2j9vfghu	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	NcGridViewV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	15	2022-05-18 11:24:12.982747+02	2022-05-18 11:24:12.982747+02
cl_2jgt0lwyo55epa	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	SuppliersRead	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	13	2022-05-18 11:24:13.062943+02	2022-05-18 11:24:13.062943+02
cl_i8optm3oaj6xqp	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	NcViewsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	19	2022-05-18 11:24:13.067211+02	2022-05-18 11:24:13.067211+02
cl_t3dlbbwzew7sam	ds_aej9ao4mzw9cff	p_41297240e6of48	md_j409llgiqo90yz	EmployeeTerritoriesList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	4	2022-05-18 11:24:12.864961+02	2022-05-18 11:24:13.28863+02
cl_7ch3641kd324n7	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	13	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.664041+02	2022-05-18 11:24:12.664041+02
cl_6f3982cef3d446	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	ErrorMessage	error_message	SingleLineText	character varying	\N	\N	255	13	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.667287+02	2022-05-18 11:24:12.667287+02
cl_pnrrjhwfpyqi12	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	Category	category	SingleLineText	character varying	\N	\N	255	13	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.669556+02	2022-05-18 11:24:12.669556+02
cl_1cgi8am13r1fao	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Url	url	LongText	text	\N	\N	\N	13	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.67543+02	2022-05-18 11:24:12.67543+02
cl_kd0zz03498jx5q	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	EmailVerificationToken	email_verification_token	SingleLineText	character varying	\N	\N	255	13	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	13	2022-05-18 11:24:12.680553+02	2022-05-18 11:24:12.680553+02
cl_6smvoxg4vqvehd	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	Error	error	LongText	text	\N	\N	\N	14	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.695128+02	2022-05-18 11:24:12.695128+02
cl_f41hgjkayi76ap	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	14	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.703958+02	2022-05-18 11:24:12.703958+02
cl_sn5wo6n2vcf8tr	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Retries	retries	Number	integer	32	0	\N	17	f	\N	f	f	\N	f	\N	0	\N	\N	integer	32	0	f	\N	\N	\N	f	17	2022-05-18 11:24:12.758297+02	2022-05-18 11:24:12.758297+02
cl_rwie4in2uxuz1i	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	RetryInterval	retry_interval	Number	integer	32	0	\N	18	f	\N	f	f	\N	f	\N	60000	\N	\N	integer	32	0	f	\N	\N	\N	f	18	2022-05-18 11:24:12.772669+02	2022-05-18 11:24:12.772669+02
cl_w0y5qxjh7p18ba	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Timeout	timeout	Number	integer	32	0	\N	19	f	\N	f	f	\N	f	\N	60000	\N	\N	integer	32	0	f	\N	\N	\N	f	19	2022-05-18 11:24:12.782626+02	2022-05-18 11:24:12.782626+02
cl_aoi05i0khwxm5g	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Active	active	Checkbox	boolean	\N	\N	\N	20	f	\N	f	f	\N	f	\N	true	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	20	2022-05-18 11:24:12.787157+02	2022-05-18 11:24:12.787157+02
cl_xej197pp067b4h	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	21	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	21	2022-05-18 11:24:12.792598+02	2022-05-18 11:24:12.792598+02
cl_6iz63txm4hfh7f	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	22	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	22	2022-05-18 11:24:12.798133+02	2022-05-18 11:24:12.798133+02
cl_318j4uypmgfmdn	ds_aej9ao4mzw9cff	p_41297240e6of48	md_zmhiordb5vpwkf	NcGridViewColumnsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	7	2022-05-18 11:24:12.866651+02	2022-05-18 11:24:12.866651+02
cl_jmagj4k4jx5ffw	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	NcAuditV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	12	2022-05-18 11:24:12.871233+02	2022-05-18 11:24:12.871233+02
cl_u1wxcvmwmn6urv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	NcFilterExpV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	15	2022-05-18 11:24:12.873743+02	2022-05-18 11:24:12.873743+02
cl_msju6cxj8nqjm4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_j409llgiqo90yz	RegionRead	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	5	2022-05-18 11:24:12.966874+02	2022-05-18 11:24:12.966874+02
cl_nbodzf77l55cru	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	OrdersList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	13	2022-05-18 11:24:12.972914+02	2022-05-18 11:24:12.972914+02
cl_74pt8m8ck2391p	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	NcFormViewV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	16	2022-05-18 11:24:12.980934+02	2022-05-18 11:24:12.980934+02
cl_m4ojzvyzshvizr	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	CustomersRead	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	16	2022-05-18 11:24:12.985284+02	2022-05-18 11:24:12.985284+02
cl_uule6x2nkxw84p	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	NcModelsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	24	2022-05-18 11:24:12.987734+02	2022-05-18 11:24:12.987734+02
cl_ed9em6bn75thju	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	NcModelsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	18	2022-05-18 11:24:13.056748+02	2022-05-18 11:24:13.056748+02
cl_j5ssutoh7jpsjp	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColLookupV2List2	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	37	2022-05-18 11:24:13.101935+02	2022-05-18 11:24:13.101935+02
cl_wjjizc3l97gv2y	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColRelationsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	38	2022-05-18 11:24:13.122697+02	2022-05-18 11:24:13.122697+02
cl_nhb4fulxib5qgn	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColRelationsV2List1	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	39	2022-05-18 11:24:13.136884+02	2022-05-18 11:24:13.136884+02
cl_ll6jzhxjga93yn	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColRelationsV2List2	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	40	2022-05-18 11:24:13.152091+02	2022-05-18 11:24:13.152091+02
cl_nht4pi4vgcjvr2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColRelationsV2List3	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	41	2022-05-18 11:24:13.167499+02	2022-05-18 11:24:13.167499+02
cl_s55zpov56wgfmh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColRelationsV2List4	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	42	2022-05-18 11:24:13.174274+02	2022-05-18 11:24:13.174274+02
cl_200i4b3amjhxyg	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColRollupV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	43	2022-05-18 11:24:13.18233+02	2022-05-18 11:24:13.18233+02
cl_stxhadl7xlmpqv	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColRollupV2List1	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	44	2022-05-18 11:24:13.188197+02	2022-05-18 11:24:13.188197+02
cl_meczg88bpm9dk2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColRollupV2List2	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	45	2022-05-18 11:24:13.193565+02	2022-05-18 11:24:13.193565+02
cl_s0x15kjbe4c2at	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColSelectOptionsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	46	2022-05-18 11:24:13.200928+02	2022-05-18 11:24:13.200928+02
cl_et3x24qyhtam8z	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcFilterExpV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	47	2022-05-18 11:24:13.206279+02	2022-05-18 11:24:13.206279+02
cl_iaf4ewburqa3fd	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcFormViewColumnsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	48	2022-05-18 11:24:13.213355+02	2022-05-18 11:24:13.213355+02
cl_ob4dqzwkr599rk	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcGalleryViewColumnsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	49	2022-05-18 11:24:13.219629+02	2022-05-18 11:24:13.219629+02
cl_83lo23d61k9mdi	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcGalleryViewV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	50	2022-05-18 11:24:13.225761+02	2022-05-18 11:24:13.225761+02
cl_h1zilufq67bh99	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcGridViewColumnsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	51	2022-05-18 11:24:13.233332+02	2022-05-18 11:24:13.233332+02
cl_lbsn3zcyhy5knh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcKanbanViewColumnsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	52	2022-05-18 11:24:13.239246+02	2022-05-18 11:24:13.239246+02
cl_d13oki8bjyop85	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcSortV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	53	2022-05-18 11:24:13.246524+02	2022-05-18 11:24:13.246524+02
cl_kh8758rfbidubl	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcModelsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	54	2022-05-18 11:24:13.252965+02	2022-05-18 11:24:13.252965+02
cl_xzi1g0bfxea38l	ds_aej9ao4mzw9cff	p_41297240e6of48	md_71j5r6adaq9l3i	CustomersMMList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	2022-05-18 11:24:13.258926+02	2022-05-18 11:24:13.258926+02
cl_xvgp8vm7xginj2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	CustomerDemographicsMMList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	2022-05-18 11:24:13.26526+02	2022-05-18 11:24:13.26526+02
cl_h44rjic2ztmxeb	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	EmployeeTerritoriesList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	19	2022-05-18 11:24:12.876984+02	2022-05-18 11:24:13.287649+02
cl_122dzvsqblwsqd	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	InputSchema	input_schema	LongText	text	\N	\N	\N	14	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.700131+02	2022-05-18 11:24:12.700131+02
cl_ukn6ghqw1xr48j	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	ResponsiveColumns	responsive_columns	SingleLineText	character varying	\N	\N	255	14	f	\N	f	f	\N	f	\N	\N	\N	\N	character varying	255	\N	f	\N	\N	\N	f	14	2022-05-18 11:24:12.703276+02	2022-05-18 11:24:12.703276+02
cl_qedxiu1jd91xje	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Un	un	Checkbox	boolean	\N	\N	\N	16	f	\N	f	f	\N	f	\N	\N	\N	\N	boolean	\N	\N	f	\N	\N	\N	f	16	2022-05-18 11:24:12.738869+02	2022-05-18 11:24:12.738869+02
cl_pc6ytykx2z82a6	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	Notification	notification	LongText	text	\N	\N	\N	16	f	\N	f	f	\N	f	\N	\N	\N	\N	text	\N	\N	f	\N	\N	\N	f	16	2022-05-18 11:24:12.743779+02	2022-05-18 11:24:12.743779+02
cl_k8gnkl4qn9ylb2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_0z5ofu8o9xph1j	NcOrgsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	6	2022-05-18 11:24:12.865804+02	2022-05-18 11:24:12.865804+02
cl_95l3l24r682gvn	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	NcColumnsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	14	2022-05-18 11:24:12.873206+02	2022-05-18 11:24:12.873206+02
cl_hsqf01u6x6uwz7	ds_aej9ao4mzw9cff	p_41297240e6of48	md_nj2q6tfni3b5a8	NcTeamsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	6	2022-05-18 11:24:12.971982+02	2022-05-18 11:24:12.971982+02
cl_uqphjc9xpcdo96	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	EmployeesList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	20	2022-05-18 11:24:12.983418+02	2022-05-18 11:24:12.983418+02
cl_8cvwgubyxtqrep	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColLookupV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	35	2022-05-18 11:24:12.987179+02	2022-05-18 11:24:12.987179+02
cl_q3y938j1f63exf	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	NcHooksV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	17	2022-05-18 11:24:13.0637+02	2022-05-18 11:24:13.0637+02
cl_qwyxc7u35tff2f	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	19	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	19	2022-05-18 11:24:12.781738+02	2022-05-18 11:24:12.781738+02
cl_dic1zbej0ei7x4	ds_aej9ao4mzw9cff	p_41297240e6of48	md_jgq6gbzvyfbk2g	TerritoriesList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	3	2022-05-18 11:24:12.864555+02	2022-05-18 11:24:12.864555+02
cl_6fou6htyzx6fa8	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5hjnvbjhgfiz6r	NcColumnsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	8	2022-05-18 11:24:12.869222+02	2022-05-18 11:24:12.869222+02
cl_7s13f8gdnr3s0p	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	NcColumnsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	13	2022-05-18 11:24:12.871798+02	2022-05-18 11:24:12.871798+02
cl_wzx2z9fw9kc75l	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	OrderDetailsList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	15	2022-05-18 11:24:12.874302+02	2022-05-18 11:24:12.874302+02
cl_9rx74ga8teuo7x	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	NcFilterExpV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	23	2022-05-18 11:24:12.878617+02	2022-05-18 11:24:12.878617+02
cl_38ziv4dkkdh8iw	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	NcViewsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	14	2022-05-18 11:24:12.974165+02	2022-05-18 11:24:12.974165+02
cl_5szjsumbk0yl6e	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	NcBasesV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	17	2022-05-18 11:24:12.981471+02	2022-05-18 11:24:12.981471+02
cl_l7p1sgvsfxg2m2	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	NcViewsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	18	2022-05-18 11:24:12.985801+02	2022-05-18 11:24:12.985801+02
cl_5o0i7pgrh2i5oh	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	NcColumnsV2Read2	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	21	2022-05-18 11:24:13.065766+02	2022-05-18 11:24:13.065766+02
cl_xnqmpmsx7xxspc	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	NcProjectUsersV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	19	2022-05-18 11:24:13.095392+02	2022-05-18 11:24:13.095392+02
cl_nweuxln7xxbqdz	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	Order	order	Decimal	real	24	\N	\N	31	f	\N	f	f	\N	f	\N	\N	\N	\N	real	24	\N	f	\N	\N	\N	f	31	2022-05-18 11:24:12.836416+02	2022-05-18 11:24:12.836416+02
cl_17imylhradkyjr	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	CreatedAt	created_at	DateTime	timestamp with time zone	\N	\N	\N	32	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	32	2022-05-18 11:24:12.839629+02	2022-05-18 11:24:12.839629+02
cl_gn8fy43v6ost0s	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	UpdatedAt	updated_at	DateTime	timestamp with time zone	\N	\N	\N	33	f	\N	t	f	\N	f	\N	CURRENT_TIMESTAMP	\N	\N	timestamp with time zone	6	\N	f	\N	\N	\N	f	33	2022-05-18 11:24:12.84279+02	2022-05-18 11:24:12.84279+02
cl_tbvbq12n5xrtso	ds_aej9ao4mzw9cff	p_41297240e6of48	md_7vpsxklk3sph75	CustomerDemographicsRead	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	3	2022-05-18 11:24:12.863789+02	2022-05-18 11:24:12.863789+02
cl_nqzdipc2k8fwu9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_s9jny9njrhdnwz	OrdersRead	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	6	2022-05-18 11:24:12.86623+02	2022-05-18 11:24:12.86623+02
cl_d0if632wlv7hiu	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	NcBasesV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	15	2022-05-18 11:24:12.873454+02	2022-05-18 11:24:12.873454+02
cl_9q88yelvrwjb6u	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	NcAuditV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	17	2022-05-18 11:24:12.876491+02	2022-05-18 11:24:12.876491+02
cl_n13v5fxz7h0dgn	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	NcModelsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	13	2022-05-18 11:24:12.980052+02	2022-05-18 11:24:12.980052+02
cl_qdrhybis5cbz4b	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	NcColumnsV2Read1	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	20	2022-05-18 11:24:12.984889+02	2022-05-18 11:24:12.984889+02
cl_xvb4bw8js0m928	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	NcColumnsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	18	2022-05-18 11:24:12.987386+02	2022-05-18 11:24:12.987386+02
cl_il02qiiyaybb14	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	NcProjectsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	14	2022-05-18 11:24:13.056289+02	2022-05-18 11:24:13.056289+02
cl_kwcxljamy9qed3	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	NcColRelationsV2List1	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	19	2022-05-18 11:24:13.064803+02	2022-05-18 11:24:13.064803+02
cl_s0w0vr1psqp3jm	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	NcColumnsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	20	2022-05-18 11:24:13.100818+02	2022-05-18 11:24:13.100818+02
cl_maydm3rcbofz04	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	NcHooksV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	21	2022-05-18 11:24:13.121596+02	2022-05-18 11:24:13.121596+02
cl_p1fuzr7pw5w24g	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	NcViewsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	22	2022-05-18 11:24:13.135105+02	2022-05-18 11:24:13.135105+02
cl_vo3yhpdpc37tle	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	NcBasesV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	23	2022-05-18 11:24:13.146578+02	2022-05-18 11:24:13.146578+02
cl_nv3xs219vjcf81	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	NcProjectsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	24	2022-05-18 11:24:13.157069+02	2022-05-18 11:24:13.157069+02
cl_go4hvqdi86lz0l	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	CustomerCustomerDemoList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	12	2022-05-18 11:24:12.870939+02	2022-05-18 11:24:13.272555+02
cl_jtqvujcd57ffc1	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	NcGridViewV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	20	2022-05-18 11:24:13.131331+02	2022-05-18 11:24:13.131331+02
cl_7yabknvt1yvb2p	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	NcKanbanViewV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	21	2022-05-18 11:24:13.148045+02	2022-05-18 11:24:13.148045+02
cl_87tqdqa6v32q7t	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	NcSharedViewsV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	22	2022-05-18 11:24:13.165399+02	2022-05-18 11:24:13.165399+02
cl_vbpxe8p4sakrfa	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	NcSortV2List	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	23	2022-05-18 11:24:13.172752+02	2022-05-18 11:24:13.172752+02
cl_zc7rvd1ff56cib	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	NcModelsV2Read	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	24	2022-05-18 11:24:13.180761+02	2022-05-18 11:24:13.180761+02
cl_9w6yddj9cco5ke	ds_aej9ao4mzw9cff	p_41297240e6of48	md_71j5r6adaq9l3i	CustomerCustomerDemoList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	3	2022-05-18 11:24:12.864037+02	2022-05-18 11:24:13.271444+02
cl_n0gu1q86ugzp2h	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	TerritoriesMMList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	2022-05-18 11:24:13.274883+02	2022-05-18 11:24:13.274883+02
cl_hbp7w3tjwejjua	ds_aej9ao4mzw9cff	p_41297240e6of48	md_j409llgiqo90yz	EmployeesMMList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	2022-05-18 11:24:13.281244+02	2022-05-18 11:24:13.281244+02
cl_b7z0ydp2g1nxu9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_nj2q6tfni3b5a8	NcUsersV2MMList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	2022-05-18 11:24:13.297022+02	2022-05-18 11:24:13.297022+02
cl_ezsdkn1opttaz3	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	NcOrgsV2MMList	\N	LinkToAnotherRecord	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	2022-05-18 11:24:13.302594+02	2022-05-18 11:24:13.302594+02
cl_pjm18dcg199nid	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	test	\N	Formula	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	2022-05-18 11:58:47.526611+02	2022-05-18 12:01:05.450512+02
\.


--
-- Data for Name: nc_cron; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_cron (id, project_id, db_alias, title, description, env, pattern, webhook, timezone, active, cron_handler, payload, headers, retries, retry_interval, timeout, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_disabled_models_for_role; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_disabled_models_for_role (id, project_id, db_alias, title, type, role, disabled, tn, rtn, cn, rcn, relation_type, created_at, updated_at, parent_model_title) FROM stdin;
\.


--
-- Data for Name: nc_disabled_models_for_role_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_disabled_models_for_role_v2 (id, base_id, project_id, fk_view_id, role, disabled, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_evolutions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_evolutions (id, title, "titleDown", description, batch, checksum, status, created, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_filter_exp_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_filter_exp_v2 (id, base_id, project_id, fk_view_id, fk_hook_id, fk_column_id, fk_parent_id, logical_op, comparison_op, value, is_group, "order", created_at, updated_at) FROM stdin;
fi_844co13k9vf1bf	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_29tn3xki80n7e3	\N	cl_nzd6ey7esqq626	\N	and	eq	France	\N	1	2022-05-18 11:50:43.239053+02	2022-05-18 11:50:51.045464+02
fi_13c22knd7wh00e	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	\N	cl_nzd6ey7esqq626	\N	and	eq	France	\N	1	2022-05-18 12:02:23.17705+02	2022-05-18 12:02:23.17705+02
\.


--
-- Data for Name: nc_form_view_columns_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_form_view_columns_v2 (id, base_id, project_id, fk_view_id, fk_column_id, uuid, label, help, description, required, show, "order", created_at, updated_at) FROM stdin;
fvc_igj3n7ghq7u8qc	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_lwxz8ubw0dlz2l	\N	\N	\N	\N	\N	t	1	2022-05-18 12:02:23.180529+02	2022-05-18 12:02:23.180529+02
fvc_vc36zaozr6s9gj	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_48210hus0kkb2v	\N	\N	\N	\N	\N	f	2	2022-05-18 12:02:23.183561+02	2022-05-18 12:02:23.183561+02
fvc_17gqudw3y7e00h	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_2n98v3z6z9y6eo	\N	\N	\N	\N	\N	f	3	2022-05-18 12:02:23.185577+02	2022-05-18 12:02:23.185577+02
fvc_hcd40a5hhm81m6	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_42mxp318j7y76i	\N	\N	\N	\N	\N	t	4	2022-05-18 12:02:23.189387+02	2022-05-18 12:02:23.189387+02
fvc_rw763dv3nv4fgv	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_h0pdckpzut20m0	\N	\N	\N	\N	\N	t	5	2022-05-18 12:02:23.192832+02	2022-05-18 12:02:23.192832+02
fvc_9fgfzckrfh8g7i	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_min1a7d8k76n5i	\N	\N	\N	\N	\N	f	6	2022-05-18 12:02:23.194853+02	2022-05-18 12:02:23.194853+02
fvc_tgzk0gkdsevbt7	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_lxtgajo4s3br49	\N	\N	\N	\N	\N	f	7	2022-05-18 12:02:23.196785+02	2022-05-18 12:02:23.196785+02
fvc_ilxc4km7v79d55	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_wo85353duh2q82	\N	\N	\N	\N	\N	f	8	2022-05-18 12:02:23.199639+02	2022-05-18 12:02:23.199639+02
fvc_c2dq5vehoxlarz	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_2hvbzdhayp3m4d	\N	\N	\N	\N	\N	f	9	2022-05-18 12:02:23.201563+02	2022-05-18 12:02:23.201563+02
fvc_hbr5v3qc4uiaru	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_n2appwjzrymdk5	\N	\N	\N	\N	\N	f	10	2022-05-18 12:02:23.203601+02	2022-05-18 12:02:23.203601+02
fvc_h53hb8tof7gt5n	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_ej7h0ko9o7l40k	\N	\N	\N	\N	\N	f	11	2022-05-18 12:02:23.207419+02	2022-05-18 12:02:23.207419+02
fvc_motdfdb71r0bgi	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_k2zpyuzl42o0ec	\N	\N	\N	\N	\N	f	12	2022-05-18 12:02:23.209777+02	2022-05-18 12:02:23.209777+02
fvc_v37rhlwnrnpt9c	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_wzx2z9fw9kc75l	\N	\N	\N	\N	\N	f	15	2022-05-18 12:02:23.216568+02	2022-05-18 12:02:23.216568+02
fvc_arfjf1r6sdt9h0	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_yl6pabbkd8b0yk	\N	\N	\N	\N	\N	f	18	2022-05-18 12:02:23.224566+02	2022-05-18 12:02:23.224566+02
fvc_jsxfottyxxce6i	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_bsjstv693k5gdj	\N	\N	\N	\N	\N	t	3	2022-05-18 12:02:23.212091+02	2022-05-18 12:02:41.570426+02
fvc_dxwt1pd81zs52z	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_nzd6ey7esqq626	\N	\N	\N	\N	\N	t	2.5	2022-05-18 12:02:23.214401+02	2022-05-18 12:02:46.228328+02
fvc_pz5pi5fb8pelg7	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_m4ojzvyzshvizr	\N	\N	\N	\N	\N	t	20	2022-05-18 12:02:23.218599+02	2022-05-18 12:04:19.541236+02
fvc_qjupuyj6ua6c5j	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_03jw5hlwoiw2ji	\N	\N	\N	\N	\N	t	2	2022-05-18 12:02:23.221226+02	2022-05-18 12:04:21.717123+02
fvc_fxfua5lmon1i9v	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_pjm18dcg199nid	\N	\N	\N	\N	\N	f	19	2022-05-18 12:02:23.226971+02	2022-05-18 12:05:17.253827+02
\.


--
-- Data for Name: nc_form_view_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_form_view_v2 (base_id, project_id, fk_view_id, heading, subheading, success_msg, redirect_url, redirect_after_secs, email, submit_another_form, show_blank_form, uuid, banner_image_url, logo_url, created_at, updated_at) FROM stdin;
ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	Input Orders	this is a test input form	\N	\N	\N	\N	\N	\N	\N	\N	\N	2022-05-18 12:02:23.167326+02	2022-05-18 12:05:46.907395+02
\.


--
-- Data for Name: nc_gallery_view_columns_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_gallery_view_columns_v2 (id, base_id, project_id, fk_view_id, fk_column_id, uuid, label, help, show, "order", created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_gallery_view_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_gallery_view_v2 (base_id, project_id, fk_view_id, next_enabled, prev_enabled, cover_image_idx, fk_cover_image_col_id, cover_image, restrict_types, restrict_size, restrict_number, public, dimensions, responsive_columns, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_grid_view_columns_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_grid_view_columns_v2 (id, fk_view_id, fk_column_id, base_id, project_id, uuid, label, help, width, show, "order", created_at, updated_at) FROM stdin;
nc_shdpu139pgu60h	vw_om33x8ttx67est	cl_gpkukq2p7xpr9n	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.840895+02	2022-05-18 11:24:11.840895+02
nc_fq1n3bvj48fv6a	vw_xfg9ot5ntsq77q	cl_jta9jl2qws2s1t	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.845081+02	2022-05-18 11:24:11.845081+02
nc_qljtnonxr58dr5	vw_c525afyrdh8nqn	cl_9wjh8ccwe2bpmz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.85015+02	2022-05-18 11:24:11.85015+02
nc_t8vjt8tocroorg	vw_tsu33qayws91ma	cl_bq67f15udwlp8d	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.857428+02	2022-05-18 11:24:11.857428+02
nc_do6ms19f6tw6dy	vw_mxy2on3nt3kj0g	cl_k5ansjw5dhsxmw	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.867747+02	2022-05-18 11:24:11.867747+02
nc_9j29o22lmwjngv	vw_ipcsobcohj1b6c	cl_t9q9o6itn46bdj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.879126+02	2022-05-18 11:24:11.879126+02
nc_uatgcpmtybbrwh	vw_h61l96me3nfmwj	cl_mgbs2x2eusxqn9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.882034+02	2022-05-18 11:24:11.882034+02
nc_4lsahdr90tmscq	vw_u2jzjq9smernol	cl_99h5zfuj8v82zf	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.883139+02	2022-05-18 11:24:11.883139+02
nc_hflgl24d0hwkb6	vw_jq0zcol66s3voo	cl_ydjxav6o2pfyt2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.883611+02	2022-05-18 11:24:11.883611+02
nc_snrjotf03mf51d	vw_2lgkfm9g3xygp1	cl_m85rflkcl4gkms	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.883899+02	2022-05-18 11:24:11.883899+02
nc_ellwfwk9z2pglh	vw_dzgaxceud97gqt	cl_q8fyky7flb1b4w	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.884593+02	2022-05-18 11:24:11.884593+02
nc_st64g80tz45fty	vw_jbveqkpzj241da	cl_jccpvuyca95upl	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.88503+02	2022-05-18 11:24:11.88503+02
nc_jhcqsk08ev4357	vw_gl8cwxr6m3iq38	cl_zkzu34hln237a9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.885322+02	2022-05-18 11:24:11.885322+02
nc_pq5bd75psvsknu	vw_ewjjjsi3mbc5h9	cl_5k0aee0pi9qjra	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.886199+02	2022-05-18 11:24:11.886199+02
nc_874ichdb62imuw	vw_sxajywo4jx9ml5	cl_pzpkxfvrsf6mbq	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.888513+02	2022-05-18 11:24:11.888513+02
nc_8m85suzkp7cz56	vw_pcdpz0lqr00s9k	cl_1947g4olf8l61e	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.889358+02	2022-05-18 11:24:11.889358+02
nc_3y5dqt59xhh970	vw_etd91poe0pvyld	cl_haf6wz56kfhhoz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.890799+02	2022-05-18 11:24:11.890799+02
nc_aenjyfn4gqpw59	vw_hk5zatfifol9yl	cl_yuc04l9uriu78e	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.891397+02	2022-05-18 11:24:11.891397+02
nc_9gyqlsi429l2dv	vw_wflzr4fus2ngby	cl_1fvruzuc9dq17k	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.894242+02	2022-05-18 11:24:11.894242+02
nc_iobewb18vyxn2y	vw_andyomdmvq8tv6	cl_c7u4guohbetlzu	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.895157+02	2022-05-18 11:24:11.895157+02
nc_z9j2mm1rnan6o1	vw_avtdalhk6frgn9	cl_gtefjqnxocsm95	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.896847+02	2022-05-18 11:24:11.896847+02
nc_h9qnqi2k51g20w	vw_6exoiii01n21je	cl_7hk845mxxfw9nb	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.899163+02	2022-05-18 11:24:11.899163+02
nc_lm2n9ywdpvfocq	vw_7fueohglirse69	cl_c0p5yzzp5h2fdv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.900207+02	2022-05-18 11:24:11.900207+02
nc_uzazqniv8z1cya	vw_n362p00huzpsod	cl_py299lng6pd9y4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.90146+02	2022-05-18 11:24:11.90146+02
nc_8c4dyzeheryeew	vw_go1u33h41idruk	cl_kd7e1qpculbik0	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.902041+02	2022-05-18 11:24:11.902041+02
nc_fywqle3cpdhn6z	vw_ea8mjggb90pnnt	cl_zdgsrgtaf7i8jh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.904383+02	2022-05-18 11:24:11.904383+02
nc_hjr9ecdulvgilz	vw_lalt4t2bjas0gj	cl_f1vep41nkme5bs	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.904914+02	2022-05-18 11:24:11.904914+02
nc_ewyj5pp7n5lvb4	vw_352nd6p9zohj5w	cl_actvro2vfybjau	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.907418+02	2022-05-18 11:24:11.907418+02
nc_4vjogxba1v6gzp	vw_7ppnzs3y6eip9i	cl_k8wnnx3bnzcy95	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.908549+02	2022-05-18 11:24:11.908549+02
nc_llani0a7vsq4dz	vw_bg9irqfa68gtl3	cl_eqywr4zelb2x2k	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.910217+02	2022-05-18 11:24:11.910217+02
nc_l15ew0umsjiaqn	vw_q1cl9wv3hhuvca	cl_pvlhgpc2mbanhq	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.911696+02	2022-05-18 11:24:11.911696+02
nc_rsoo7om1rfeswx	vw_lr57pe50bbmjar	cl_eztxs6bu5hddmh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.913241+02	2022-05-18 11:24:11.913241+02
nc_qntzlktm5cx118	vw_mauaaqk5b7v0rr	cl_gnsoew7es07l63	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.914232+02	2022-05-18 11:24:11.914232+02
nc_d47zwiktiawpvt	vw_t8o4jwxtrx9nf0	cl_hl2a11v360e4b1	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.914539+02	2022-05-18 11:24:11.914539+02
nc_gohhk2uz0l3363	vw_51d0c7bmbclwb8	cl_mhh2pb9uvhqu8d	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.917409+02	2022-05-18 11:24:11.917409+02
nc_snkg98obvuh9l8	vw_brn7rul3d130ex	cl_duxrw3cakjkz2l	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.921242+02	2022-05-18 11:24:11.921242+02
nc_t8wzbo6mqchoq0	vw_dejssdomuxg3uw	cl_1sizugbjggamtt	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.923515+02	2022-05-18 11:24:11.923515+02
nc_ysbs6lnp7d1t84	vw_du7zrxhj51z27u	cl_vvtlvtc6gzirw0	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.93281+02	2022-05-18 11:24:11.93281+02
nc_vovrbfc03yk6uz	vw_5w6mg5ux0ryjjl	cl_ji20zt281pec2c	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.935339+02	2022-05-18 11:24:11.935339+02
nc_22o8917iqxd16x	vw_srnlvrf0gn6rul	cl_hflevolpf863c2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.9361+02	2022-05-18 11:24:11.9361+02
nc_6rv7y95y175x7u	vw_ccggsien3mx2nl	cl_z4jql0g2ge0i9x	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.937263+02	2022-05-18 11:24:11.937263+02
nc_ews6zcfgyk4o7s	vw_r15acq81tqlkxp	cl_0x0h3rq3ygplwc	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.938832+02	2022-05-18 11:24:11.938832+02
nc_452pr6fwwmwp4r	vw_81imke2v3e15l5	cl_6i16gsjxn2t5bh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.939316+02	2022-05-18 11:24:11.939316+02
nc_uf16k1oagpga7j	vw_idcyonon9t2t5l	cl_u7wce2nao3rabj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.940091+02	2022-05-18 11:24:11.940091+02
nc_jk7bqunhy6fwaa	vw_yrmq5jknxj9sx6	cl_51c81maycrxne7	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.940393+02	2022-05-18 11:24:11.940393+02
nc_517qts1nw3eog2	vw_om33x8ttx67est	cl_yk84aiwkkjj5fq	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.948321+02	2022-05-18 11:24:11.948321+02
nc_t3xwo0xja1mtr8	vw_xfg9ot5ntsq77q	cl_7vy94bkvan3b13	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.951696+02	2022-05-18 11:24:11.951696+02
nc_o121y8mxkpymbx	vw_8kwir8e38t3fc9	cl_dvr9xeexoglp9b	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.954011+02	2022-05-18 11:24:11.954011+02
nc_uus9iy8mfh3vro	vw_c525afyrdh8nqn	cl_px1s562w4pnn24	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.958842+02	2022-05-18 11:24:11.958842+02
nc_janwrhww889l4j	vw_tsu33qayws91ma	cl_viup64dwqpa7qw	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.963027+02	2022-05-18 11:24:11.963027+02
nc_1e5shq5w7zelcn	vw_29tn3xki80n7e3	cl_lwxz8ubw0dlz2l	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	1	2022-05-18 11:24:11.929154+02	2022-05-18 11:56:40.632476+02
nc_466uk447j3rl7r	vw_mxy2on3nt3kj0g	cl_0qe00wlxhp6l8d	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.974808+02	2022-05-18 11:24:11.974808+02
nc_xvsmxq2czz2ic7	vw_u2jzjq9smernol	cl_p7fzwqrfnshgt5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.98631+02	2022-05-18 11:24:11.98631+02
nc_1kkbpwcsy1ypgc	vw_jq0zcol66s3voo	cl_bg6e57ixpkc66e	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.99224+02	2022-05-18 11:24:11.99224+02
nc_x6p2wgd2gro3vb	vw_lalt4t2bjas0gj	cl_ijd8a2zvsvzxdb	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.006877+02	2022-05-18 11:24:12.006877+02
nc_vwhnb2v9rc2qq5	vw_lr57pe50bbmjar	cl_2vsljvxzziveo3	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.015728+02	2022-05-18 11:24:12.015728+02
nc_e6n6huq0e9f0ip	vw_t8o4jwxtrx9nf0	cl_o1e6spelqpafwg	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.01968+02	2022-05-18 11:24:12.01968+02
nc_i1eo0rjh5u4pk6	vw_ccggsien3mx2nl	cl_u6d2985s4k2vje	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.037075+02	2022-05-18 11:24:12.037075+02
nc_26aw8dayk65vay	vw_pcdpz0lqr00s9k	cl_8rpusw386qct5p	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.08693+02	2022-05-18 11:24:12.08693+02
nc_uoc4wzvo8twh8a	vw_6exoiii01n21je	cl_yldbutsud8s8es	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.103955+02	2022-05-18 11:24:12.103955+02
nc_nu5a43gwoyctzg	vw_51d0c7bmbclwb8	cl_cp91gcskjlxt5k	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.107184+02	2022-05-18 11:24:12.107184+02
nc_2i2jbk0cgu18gn	vw_81imke2v3e15l5	cl_fifa1pn2w5hhiv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.134203+02	2022-05-18 11:24:12.134203+02
nc_aezq0y6xb34vae	vw_jq0zcol66s3voo	cl_roh6z8p6y9kt74	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.172537+02	2022-05-18 11:24:12.172537+02
nc_9fahullv5hrnyg	vw_wflzr4fus2ngby	cl_77iddmy4xahfwa	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.175219+02	2022-05-18 11:24:12.175219+02
nc_l5j5ilh5agwes9	vw_t8o4jwxtrx9nf0	cl_2etavxey6fg0fo	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.194398+02	2022-05-18 11:24:12.194398+02
nc_8ju95g3u37ca02	vw_dejssdomuxg3uw	cl_ove7ejfffgx03c	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.20246+02	2022-05-18 11:24:12.20246+02
nc_yfoymbhli7m7vj	vw_dejssdomuxg3uw	cl_3sv6p41h8dk62h	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.272506+02	2022-05-18 11:24:12.272506+02
nc_8abks13vbim8ph	vw_srnlvrf0gn6rul	cl_ulp62fxwdxscyv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.280608+02	2022-05-18 11:24:12.280608+02
nc_hk62y12inwxy1h	vw_tsu33qayws91ma	cl_b8pnbjqcsfv8yc	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.297397+02	2022-05-18 11:24:12.297397+02
nc_0iir3f8itigq8x	vw_352nd6p9zohj5w	cl_m2jfcpnqk7qb4p	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.33252+02	2022-05-18 11:24:12.33252+02
nc_m8yuabbda2a7vo	vw_t8o4jwxtrx9nf0	cl_b4f6kuyhpgnxf6	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.337107+02	2022-05-18 11:24:12.337107+02
nc_n72b241l3gfg5b	vw_tsu33qayws91ma	cl_jd7stjnewfms0c	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.373262+02	2022-05-18 11:24:12.373262+02
nc_kmxtf2vsq4xibo	vw_pcdpz0lqr00s9k	cl_orefgn2y3dot57	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.394801+02	2022-05-18 11:24:12.394801+02
nc_w0thoddsknwhz5	vw_go1u33h41idruk	cl_0t5gqpzq3rmbub	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.407058+02	2022-05-18 11:24:12.407058+02
nc_tj5owrffw4l2ab	vw_etd91poe0pvyld	cl_r4ephzb7n37bj1	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.458135+02	2022-05-18 11:24:12.458135+02
nc_siq4hvfo08egas	vw_6exoiii01n21je	cl_jny174p7hh10qk	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.472366+02	2022-05-18 11:24:12.472366+02
nc_wj1ordc21y63xc	vw_brn7rul3d130ex	cl_u1ygzte476l7jn	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.475067+02	2022-05-18 11:24:12.475067+02
nc_634m9dlhezrpue	vw_pcdpz0lqr00s9k	cl_mz44l54qkdvw8m	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.512245+02	2022-05-18 11:24:12.512245+02
nc_8g0z3r25gjmf8t	vw_ewjjjsi3mbc5h9	cl_dp5y1a7pcz0s54	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.5181+02	2022-05-18 11:24:12.5181+02
nc_zkmavklsbd84tl	vw_lalt4t2bjas0gj	cl_9dhvtte5npaada	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.571713+02	2022-05-18 11:24:12.571713+02
nc_9gcksdma3yih53	vw_jbveqkpzj241da	cl_f4nw9fnghi0wi2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.610339+02	2022-05-18 11:24:12.610339+02
nc_l5lnzootadd12c	vw_ea8mjggb90pnnt	cl_wsrry9ph73bbf9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.619837+02	2022-05-18 11:24:12.619837+02
nc_wolr3dderu2hmu	vw_352nd6p9zohj5w	cl_qkxcxycznt8kxs	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.657768+02	2022-05-18 11:24:12.657768+02
nc_1i0d0jryn2zm7x	vw_avtdalhk6frgn9	cl_6f3982cef3d446	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.684486+02	2022-05-18 11:24:12.684486+02
nc_3fkdnb62ywfmbo	vw_u2jzjq9smernol	cl_dvwubuqeafa6qp	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.707851+02	2022-05-18 11:24:12.707851+02
nc_f8lsonqig4460p	vw_jbveqkpzj241da	cl_7sopjg9jqet3lv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.710671+02	2022-05-18 11:24:12.710671+02
nc_2t273jwy2cjret	vw_dejssdomuxg3uw	cl_v0swmi0cyni03y	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.720105+02	2022-05-18 11:24:12.720105+02
nc_citsp1gny3cdcr	vw_dejssdomuxg3uw	cl_xkvq5go221sd4b	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:12.755101+02	2022-05-18 11:24:12.755101+02
nc_03vmdermo7b1nx	vw_dejssdomuxg3uw	cl_ydq1ak4824ae8r	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:12.769352+02	2022-05-18 11:24:12.769352+02
nc_p0o0cy6phkjrzn	vw_ccggsien3mx2nl	cl_nqzdipc2k8fwu9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.940171+02	2022-05-18 11:24:12.940171+02
nc_rzimwewej1txc5	vw_hk5zatfifol9yl	cl_u1wxcvmwmn6urv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.946218+02	2022-05-18 11:24:12.946218+02
nc_7am7d7o1qq3hmx	vw_6exoiii01n21je	cl_z02hnsgf78dgxx	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:13.032899+02	2022-05-18 11:24:13.032899+02
nc_aosexif1cpw70l	vw_dejssdomuxg3uw	cl_hi074lxlm7gyan	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	19	2022-05-18 11:24:13.037831+02	2022-05-18 11:24:13.037831+02
nc_tfabi4sn3ep6yz	vw_ipcsobcohj1b6c	cl_rrv3kdmoalscjb	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:13.085994+02	2022-05-18 11:24:13.085994+02
nc_vtrvdjac5063g8	vw_29tn3xki80n7e3	cl_h0pdckpzut20m0	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.275587+02	2022-05-18 11:56:41.839236+02
nc_kieiotycdo3ztr	vw_29tn3xki80n7e3	cl_bsjstv693k5gdj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	13	2022-05-18 11:24:12.702458+02	2022-05-18 11:56:38.998077+02
nc_qharw1xns0eoxn	vw_29tn3xki80n7e3	cl_yl6pabbkd8b0yk	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	18	2022-05-18 11:24:13.124571+02	2022-05-18 11:56:38.998077+02
nc_cjko921to5uthj	vw_29tn3xki80n7e3	cl_42mxp318j7y76i	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.204919+02	2022-05-18 11:56:41.296911+02
nc_e978ed29icawik	vw_h61l96me3nfmwj	cl_2xu900p2fe407f	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.985389+02	2022-05-18 11:24:11.985389+02
nc_agr3uli6zril5g	vw_dzgaxceud97gqt	cl_narv6h4qk6asbn	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.988132+02	2022-05-18 11:24:11.988132+02
nc_54n4xt8nsroe9n	vw_sxajywo4jx9ml5	cl_dpxgtdswb10tkh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.991007+02	2022-05-18 11:24:11.991007+02
nc_2veaq5mulg05l8	vw_etd91poe0pvyld	cl_swryd2y3ho6c3t	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.995041+02	2022-05-18 11:24:11.995041+02
nc_5ari9aksg46hr0	vw_bg9irqfa68gtl3	cl_zodcdlchikjfqz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.014069+02	2022-05-18 11:24:12.014069+02
nc_pva9jzaj2y8vf1	vw_brn7rul3d130ex	cl_xzhpoilm6b0bmh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.021384+02	2022-05-18 11:24:12.021384+02
nc_lkzj6z4icxd8mp	vw_srnlvrf0gn6rul	cl_1msk1km2tcqyww	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.038929+02	2022-05-18 11:24:12.038929+02
nc_4j7ohljfof1qbd	vw_5w6mg5ux0ryjjl	cl_x7rpf6qkm426p4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.042263+02	2022-05-18 11:24:12.042263+02
nc_6a00khlx4d7xti	vw_avtdalhk6frgn9	cl_v28tzp0c0c0jt4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.09403+02	2022-05-18 11:24:12.09403+02
nc_junwsvtvahbq9o	vw_ea8mjggb90pnnt	cl_isb3w8xrlllgwm	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.1025+02	2022-05-18 11:24:12.1025+02
nc_3xik2gj7a0pqy9	vw_352nd6p9zohj5w	cl_0j274wxlvyo4yz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.105703+02	2022-05-18 11:24:12.105703+02
nc_xu7ah39duatpk8	vw_ccggsien3mx2nl	cl_vnjldzdp0ouqfw	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.131297+02	2022-05-18 11:24:12.131297+02
nc_03wih0fy1wchnq	vw_yrmq5jknxj9sx6	cl_b4w9b6bqoq0mj1	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.136133+02	2022-05-18 11:24:12.136133+02
nc_97tkivhpg4iqex	vw_n362p00huzpsod	cl_k5fdsl202of2rn	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.177184+02	2022-05-18 11:24:12.177184+02
nc_tv6floi3r21rst	vw_bg9irqfa68gtl3	cl_nb9fflwnh0el52	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.189356+02	2022-05-18 11:24:12.189356+02
nc_qnnf3zufdcpm5v	vw_ccggsien3mx2nl	cl_z04rfjpq8ajuaj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.214506+02	2022-05-18 11:24:12.214506+02
nc_wkelt3vli2miyk	vw_jbveqkpzj241da	cl_fikkm6avrnvalf	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.243236+02	2022-05-18 11:24:12.243236+02
nc_zfzxvjyxndpl43	vw_pcdpz0lqr00s9k	cl_6yy0dvgw2yb248	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.247438+02	2022-05-18 11:24:12.247438+02
nc_ym8qs19hceq81e	vw_ea8mjggb90pnnt	cl_73k8fe81h06itk	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.25912+02	2022-05-18 11:24:12.25912+02
nc_j7hdolt4y57xrt	vw_5w6mg5ux0ryjjl	cl_9g73625z1ka2js	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.290349+02	2022-05-18 11:24:12.290349+02
nc_fhf0xosopf3hkx	vw_pcdpz0lqr00s9k	cl_6zmi1axycbm9zs	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.318665+02	2022-05-18 11:24:12.318665+02
nc_nh66vudslow6y7	vw_go1u33h41idruk	cl_b45jar9oybu3vc	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.330978+02	2022-05-18 11:24:12.330978+02
nc_sqlswg5vuoyjx2	vw_brn7rul3d130ex	cl_hgpa783mbqeniv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.338934+02	2022-05-18 11:24:12.338934+02
nc_b3kihm0ntu33jc	vw_5w6mg5ux0ryjjl	cl_iszr4mbjyqie02	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.360413+02	2022-05-18 11:24:12.360413+02
nc_vt5jgpgru7rmok	vw_dzgaxceud97gqt	cl_nwguwudzn15f31	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.38945+02	2022-05-18 11:24:12.38945+02
nc_4wtt6z9k339mac	vw_etd91poe0pvyld	cl_pluaphsmwwqq5i	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.397119+02	2022-05-18 11:24:12.397119+02
nc_r4uervc3mxz0qd	vw_5w6mg5ux0ryjjl	cl_hmh16s2ljlm5rq	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.430111+02	2022-05-18 11:24:12.430111+02
nc_4tv5gllvdbic0m	vw_etd91poe0pvyld	cl_iggt7rotz7o8fe	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.510364+02	2022-05-18 11:24:12.510364+02
nc_npput5p7mexdfy	vw_gl8cwxr6m3iq38	cl_sc5t1gtu5agi6b	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.516858+02	2022-05-18 11:24:12.516858+02
nc_ytvwo92uhfpled	vw_go1u33h41idruk	cl_js7114p3q9t5sa	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.522254+02	2022-05-18 11:24:12.522254+02
nc_f5j4j937b3baih	vw_brn7rul3d130ex	cl_9sqwmb4k4dxhoc	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.524974+02	2022-05-18 11:24:12.524974+02
nc_t80rsysh6apw8m	vw_ea8mjggb90pnnt	cl_fv2czdtf2qgrw6	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.573383+02	2022-05-18 11:24:12.573383+02
nc_9dnxccy06qb2l3	vw_t8o4jwxtrx9nf0	cl_qiubigzgwgau18	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.576671+02	2022-05-18 11:24:12.576671+02
nc_6dx4jhiusips26	vw_5w6mg5ux0ryjjl	cl_9sh3add4j1h5tu	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.586837+02	2022-05-18 11:24:12.586837+02
nc_4607h0awzc6gep	vw_lalt4t2bjas0gj	cl_55duarfeg1agza	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.615403+02	2022-05-18 11:24:12.615403+02
nc_v8lchlu5n00wlr	vw_hk5zatfifol9yl	cl_cne8f9kkhhmz6d	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.649303+02	2022-05-18 11:24:12.649303+02
nc_x2tuipm2x2hn34	vw_go1u33h41idruk	cl_74xo3c6o7uhjsa	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.656606+02	2022-05-18 11:24:12.656606+02
nc_653ilg1bv94gsl	vw_jbveqkpzj241da	cl_qjegk9iz1niq0i	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.686092+02	2022-05-18 11:24:12.686092+02
nc_jmy3zy0q1bgeew	vw_ewjjjsi3mbc5h9	cl_4p1x15zwndeksm	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.690971+02	2022-05-18 11:24:12.690971+02
nc_fv566hpmba4a4w	vw_dejssdomuxg3uw	cl_kmwn5f21u4wyxf	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.739518+02	2022-05-18 11:24:12.739518+02
nc_y019m94ns78758	vw_andyomdmvq8tv6	cl_3phwdvhn8yfixc	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.741957+02	2022-05-18 11:24:12.741957+02
nc_u9ot0o73eoz30c	vw_ewjjjsi3mbc5h9	cl_cqgz84qtnyafgv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:12.756162+02	2022-05-18 11:24:12.756162+02
nc_97t6k7mvnrjvo5	vw_ewjjjsi3mbc5h9	cl_g0f1yyy9dmvuta	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:12.77028+02	2022-05-18 11:24:12.77028+02
nc_v6epj6awm2rfmu	vw_ewjjjsi3mbc5h9	cl_ltaq7rwb27qjmp	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	18	2022-05-18 11:24:12.781269+02	2022-05-18 11:24:12.781269+02
nc_4zjvuh5tlr8czd	vw_2lgkfm9g3xygp1	cl_mr8rz9hd7ttjwp	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.939202+02	2022-05-18 11:24:12.939202+02
nc_zlc9bgxb4r5l41	vw_go1u33h41idruk	cl_xo0892py3ndpke	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.941368+02	2022-05-18 11:24:12.941368+02
nc_pqq53rusvtayqy	vw_51d0c7bmbclwb8	cl_wrshd8cmtqyzk5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.944493+02	2022-05-18 11:24:12.944493+02
nc_80bo9f6rilolhc	vw_mauaaqk5b7v0rr	cl_3rcpv6o301oikz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.947552+02	2022-05-18 11:24:12.947552+02
nc_ykdtxjnm7tozu6	vw_h61l96me3nfmwj	cl_m2p63veyy3cmww	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	34	2022-05-18 11:24:12.949902+02	2022-05-18 11:24:12.949902+02
nc_ozh5v37zgv5mpl	vw_dejssdomuxg3uw	cl_vr6w15ow265mdm	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	18	2022-05-18 11:24:12.952796+02	2022-05-18 11:24:12.952796+02
nc_kcyhgfmhbe2vju	vw_ccggsien3mx2nl	cl_uq57oat6mrcthd	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:13.031887+02	2022-05-18 11:24:13.031887+02
nc_2o714081uo2dmi	vw_wflzr4fus2ngby	cl_leidqf2j9vfghu	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:13.034275+02	2022-05-18 11:24:13.034275+02
nc_gjmbvj2smezn6j	vw_ipcsobcohj1b6c	cl_cyde2oo9zyylv7	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.986767+02	2022-05-18 11:24:11.986767+02
nc_8w9k0rwrnryfbb	vw_mauaaqk5b7v0rr	cl_ygmn763v7bzofp	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.016347+02	2022-05-18 11:24:12.016347+02
nc_rb2yb4cuxa76ou	vw_wflzr4fus2ngby	cl_2jbc6l8uccbcxv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.092526+02	2022-05-18 11:24:12.092526+02
nc_uyogz9zeglsp5d	vw_andyomdmvq8tv6	cl_v2ll0vs8jz4501	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.100897+02	2022-05-18 11:24:12.100897+02
nc_a6zxonmrwezxf5	vw_bg9irqfa68gtl3	cl_aemppv0wnynryh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.107482+02	2022-05-18 11:24:12.107482+02
nc_sie6efxyt7c8bb	vw_srnlvrf0gn6rul	cl_02u9gqs7dleiah	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.12427+02	2022-05-18 11:24:12.12427+02
nc_yz0fgmw7u55uir	vw_om33x8ttx67est	cl_c548hgzjy1v6kv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.138277+02	2022-05-18 11:24:12.138277+02
nc_h56e7zqh2g9r1t	vw_2lgkfm9g3xygp1	cl_jprcavrmue3v7e	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.17552+02	2022-05-18 11:24:12.17552+02
nc_9w848achdzpe8e	vw_ea8mjggb90pnnt	cl_w9k4oudkzve2tu	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.188024+02	2022-05-18 11:24:12.188024+02
nc_kcq1tygchns28o	vw_mauaaqk5b7v0rr	cl_w10l8tx9zyh2wb	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.191232+02	2022-05-18 11:24:12.191232+02
nc_6d0xm00ahx6y18	vw_lr57pe50bbmjar	cl_lst1bydbilrj0f	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.195334+02	2022-05-18 11:24:12.195334+02
nc_7q20q4wgion655	vw_5w6mg5ux0ryjjl	cl_gi9vm39qv3xdin	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.216547+02	2022-05-18 11:24:12.216547+02
nc_ateu4jg1homfrc	vw_u2jzjq9smernol	cl_6fchl3afgxh7yp	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.241432+02	2022-05-18 11:24:12.241432+02
nc_frtcim5yih7vol	vw_ipcsobcohj1b6c	cl_es4sn3km17zc4e	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.246336+02	2022-05-18 11:24:12.246336+02
nc_f8nho0mv7qx48y	vw_etd91poe0pvyld	cl_c0u4gz4xsq5zqu	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.248833+02	2022-05-18 11:24:12.248833+02
nc_jt7f2kidkizsdn	vw_n362p00huzpsod	cl_9a6x2krsshjkks	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.251571+02	2022-05-18 11:24:12.251571+02
nc_4w6qsevhua6bjs	vw_ccggsien3mx2nl	cl_zgif7qh1drm329	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.285309+02	2022-05-18 11:24:12.285309+02
nc_53dhxrimna08h7	vw_ipcsobcohj1b6c	cl_b86m0u9nsa08p9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.317455+02	2022-05-18 11:24:12.317455+02
nc_36wzvxmnu5g0w1	vw_etd91poe0pvyld	cl_6fmcncpfbx7wy0	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.319835+02	2022-05-18 11:24:12.319835+02
nc_l1eko6e1x4zu49	vw_n362p00huzpsod	cl_hqmb5f2hlm8kbt	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.322451+02	2022-05-18 11:24:12.322451+02
nc_ua3asu4rpdk1xo	vw_ztmg51h4hma3hj	cl_edzzu8r2caweg4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.353057+02	2022-05-18 11:24:12.353057+02
nc_7jzqyl6acxgbwb	vw_wflzr4fus2ngby	cl_uhtznd7f8pokfx	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.395492+02	2022-05-18 11:24:12.395492+02
nc_5dr3j3u9syvdul	vw_jbveqkpzj241da	cl_ysuzqfhvklkqw2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.455309+02	2022-05-18 11:24:12.455309+02
nc_6ke8uaz8009qem	vw_n362p00huzpsod	cl_pukj4zblib5hya	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.462798+02	2022-05-18 11:24:12.462798+02
nc_io7d844dd1t6un	vw_go1u33h41idruk	cl_mk3hu3qmq6tdfj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.472717+02	2022-05-18 11:24:12.472717+02
nc_xrz1gewlzx6avc	vw_t8o4jwxtrx9nf0	cl_0tsvdxn69tvyh0	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.480456+02	2022-05-18 11:24:12.480456+02
nc_jxk040qraah7xs	vw_wflzr4fus2ngby	cl_m2e1p9fzymlhq4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.512757+02	2022-05-18 11:24:12.512757+02
nc_j73ral3n008zm8	vw_tsu33qayws91ma	cl_8y5rf4iiyqjfad	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.5469+02	2022-05-18 11:24:12.5469+02
nc_7v1zhggnsz10ze	vw_etd91poe0pvyld	cl_8kcusz9r08t1n5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.561411+02	2022-05-18 11:24:12.561411+02
nc_5s3pqoqh0y50sx	vw_h61l96me3nfmwj	cl_pthbmkbtet9gao	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.647783+02	2022-05-18 11:24:12.647783+02
nc_pkiww0p4l3eoud	vw_ewjjjsi3mbc5h9	cl_87e0ickdzoj44q	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.66298+02	2022-05-18 11:24:12.66298+02
nc_5165eu535sfhgy	vw_t8o4jwxtrx9nf0	cl_tespjgc8eqnn3u	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.666095+02	2022-05-18 11:24:12.666095+02
nc_h22fuzp3nc3rwe	vw_bg9irqfa68gtl3	cl_rdcmlig4o9fv6y	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.692718+02	2022-05-18 11:24:12.692718+02
nc_yqte9j80v55ip0	vw_etd91poe0pvyld	cl_xlbezj3jt14886	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.711446+02	2022-05-18 11:24:12.711446+02
nc_sq7bukfyd12gud	vw_ea8mjggb90pnnt	cl_s2e5qmhhb1jten	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.717653+02	2022-05-18 11:24:12.717653+02
nc_vx97ufgtmih9k7	vw_srnlvrf0gn6rul	cl_4y171k6kshi6e1	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.720385+02	2022-05-18 11:24:12.720385+02
nc_uvgjzj62r94ad1	vw_bg9irqfa68gtl3	cl_2z8e4cfd1sj65e	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.740958+02	2022-05-18 11:24:12.740958+02
nc_cw434p49n3u2bl	vw_u2jzjq9smernol	cl_s4i6nmg47h2j5p	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:12.752945+02	2022-05-18 11:24:12.752945+02
nc_0tub5yjgee8cai	vw_u2jzjq9smernol	cl_tqy34owdpe210n	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:12.767346+02	2022-05-18 11:24:12.767346+02
nc_dbdedamd1mxfac	vw_u2jzjq9smernol	cl_6o6s13d0gw5jk7	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	18	2022-05-18 11:24:12.778337+02	2022-05-18 11:24:12.778337+02
nc_d173hlx4vyv0tc	vw_mxy2on3nt3kj0g	cl_cz3l6lgwksaq2e	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.937883+02	2022-05-18 11:24:12.937883+02
nc_p8us6qpodq2o1m	vw_q1cl9wv3hhuvca	cl_4gstsln64mwssn	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.939965+02	2022-05-18 11:24:12.939965+02
nc_n50sbol1pbuxvl	vw_wflzr4fus2ngby	cl_95l3l24r682gvn	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.945963+02	2022-05-18 11:24:12.945963+02
nc_jziackwmk9413h	vw_81imke2v3e15l5	cl_msju6cxj8nqjm4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:13.032734+02	2022-05-18 11:24:13.032734+02
nc_se51tsf1pgjo93	vw_7fueohglirse69	cl_xdmlbpk1l1lxt6	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:13.035141+02	2022-05-18 11:24:13.035141+02
nc_67mi3kn5v0dt4a	vw_pcdpz0lqr00s9k	cl_74pt8m8ck2391p	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:13.037563+02	2022-05-18 11:24:13.037563+02
nc_2j3t2sy5bufdo9	vw_ea8mjggb90pnnt	cl_uule6x2nkxw84p	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	24	2022-05-18 11:24:13.039842+02	2022-05-18 11:24:13.039842+02
nc_lovbyokvs3ru8q	vw_andyomdmvq8tv6	cl_xvb4bw8js0m928	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	18	2022-05-18 11:24:13.043289+02	2022-05-18 11:24:13.043289+02
nc_th9me56q49jvch	vw_bg9irqfa68gtl3	cl_ed9em6bn75thju	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	18	2022-05-18 11:24:13.087588+02	2022-05-18 11:24:13.087588+02
nc_rm1zmxa8cr1iy0	vw_ztmg51h4hma3hj	cl_v9cnldmlymawdy	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	11	2022-05-18 11:24:12.942264+02	2022-05-18 11:30:29.910147+02
nc_0v5agq25oqlvaf	vw_29tn3xki80n7e3	cl_48210hus0kkb2v	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	2	2022-05-18 11:24:12.029756+02	2022-05-18 11:56:38.998077+02
nc_ay5uz946s1kuop	vw_gl8cwxr6m3iq38	cl_ycnsnkvz4v8akf	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.988473+02	2022-05-18 11:24:11.988473+02
nc_kh1ijp61p5m5aq	vw_avtdalhk6frgn9	cl_bwrx80dhb9qhjd	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.999344+02	2022-05-18 11:24:11.999344+02
nc_98t7gvqrriewm0	vw_352nd6p9zohj5w	cl_5l0iqbng3qv1wo	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.009476+02	2022-05-18 11:24:12.009476+02
nc_5t7ebmd0rwpy5j	vw_ztmg51h4hma3hj	cl_x801ithwdkcrx9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.0392+02	2022-05-18 11:24:12.0392+02
nc_ta70235oqqu71s	vw_om33x8ttx67est	cl_vn7qtpv2s09tkt	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.048253+02	2022-05-18 11:24:12.048253+02
nc_cy4id2dx1lgci5	vw_jbveqkpzj241da	cl_o5rwys2v3e6oyo	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.085773+02	2022-05-18 11:24:12.085773+02
nc_bnha64fn6bpfh8	vw_hk5zatfifol9yl	cl_us0qzk2ufx407a	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.088799+02	2022-05-18 11:24:12.088799+02
nc_644mlh9r0a33v8	vw_lr57pe50bbmjar	cl_h2is03dfyemdhi	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.106067+02	2022-05-18 11:24:12.106067+02
nc_jy38ab5mkgjzza	vw_t8o4jwxtrx9nf0	cl_kmwlqydz7qisje	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.108984+02	2022-05-18 11:24:12.108984+02
nc_5r90yqni29nslu	vw_jbveqkpzj241da	cl_nv2qb0bqfi14xj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.168991+02	2022-05-18 11:24:12.168991+02
nc_zo8yvdhn5j6s24	vw_brn7rul3d130ex	cl_s1ik0ltcq5ufcv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.192922+02	2022-05-18 11:24:12.192922+02
nc_6bsy7em2bci70x	vw_tsu33qayws91ma	cl_mae86z5veq16zm	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.224894+02	2022-05-18 11:24:12.224894+02
nc_43k52qadejlwio	vw_h61l96me3nfmwj	cl_nce63jcfgdcfn3	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.243842+02	2022-05-18 11:24:12.243842+02
nc_nj8en64wxndppy	vw_wflzr4fus2ngby	cl_cz33iocpqead80	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.250557+02	2022-05-18 11:24:12.250557+02
nc_umlxsaczkrp52l	vw_avtdalhk6frgn9	cl_to8frc2rpu5ar7	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.255918+02	2022-05-18 11:24:12.255918+02
nc_iq5xwp0eyesbnx	vw_wflzr4fus2ngby	cl_kuayl8wxyb2k5d	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.318922+02	2022-05-18 11:24:12.318922+02
nc_dvz51vm184ggcx	vw_jq0zcol66s3voo	cl_cbwd42kx2q4778	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.32137+02	2022-05-18 11:24:12.32137+02
nc_1bd2ljsiw58q2p	vw_avtdalhk6frgn9	cl_b8tbs19b6ei4oe	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.323827+02	2022-05-18 11:24:12.323827+02
nc_n5w0jvodbmwlp5	vw_6exoiii01n21je	cl_ocpkpwek7kd96e	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.33138+02	2022-05-18 11:24:12.33138+02
nc_1wokbanzsl0nvn	vw_ipcsobcohj1b6c	cl_lnb2rstu0dzd0s	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.392554+02	2022-05-18 11:24:12.392554+02
nc_a2nqgrrt6iau4o	vw_n362p00huzpsod	cl_pasy3zcvs1sdoi	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.400409+02	2022-05-18 11:24:12.400409+02
nc_70mxvmkgnfxa8y	vw_7ppnzs3y6eip9i	cl_8v2ynimciarahr	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.403253+02	2022-05-18 11:24:12.403253+02
nc_odizrg25pb9zgb	vw_mauaaqk5b7v0rr	cl_s4dj2g6xpllcm8	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.409006+02	2022-05-18 11:24:12.409006+02
nc_z2up5j8omsf54g	vw_brn7rul3d130ex	cl_jufntqmsiohfs3	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.414597+02	2022-05-18 11:24:12.414597+02
nc_rkckozs0oswpt4	vw_2lgkfm9g3xygp1	cl_3zglox2j8sjgrz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.457107+02	2022-05-18 11:24:12.457107+02
nc_f4m4v9fazp00ps	vw_jq0zcol66s3voo	cl_epm4ujkyxsksgv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.46021+02	2022-05-18 11:24:12.46021+02
nc_mx2ezzsmt7bd1n	vw_avtdalhk6frgn9	cl_bi5wfcj36qnihx	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.464195+02	2022-05-18 11:24:12.464195+02
nc_maxww0lbu78e3y	vw_n362p00huzpsod	cl_i2pwveeqruhjtz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.514249+02	2022-05-18 11:24:12.514249+02
nc_elehvi3cd4stk2	vw_t8o4jwxtrx9nf0	cl_dpp1unrz3q1c3g	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.530349+02	2022-05-18 11:24:12.530349+02
nc_cdo5vwbmdz9mnp	vw_wflzr4fus2ngby	cl_v7v1onya6tagss	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.563622+02	2022-05-18 11:24:12.563622+02
nc_zo5f1k7j2ruat0	vw_pcdpz0lqr00s9k	cl_0i5dworwktqo5u	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.566194+02	2022-05-18 11:24:12.566194+02
nc_boyt6je41s6r3w	vw_avtdalhk6frgn9	cl_sffsxt353w499o	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.568908+02	2022-05-18 11:24:12.568908+02
nc_8uzizmmyw3pjw9	vw_6exoiii01n21je	cl_kki4i38mfv3zw4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.57358+02	2022-05-18 11:24:12.57358+02
nc_yaqm5b7g9ex9ay	vw_ztmg51h4hma3hj	cl_k7kqyx1mo2y94k	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.58427+02	2022-05-18 11:24:12.58427+02
nc_w2p9vr45wg3eic	vw_srnlvrf0gn6rul	cl_rqltw519nd1bvk	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.587149+02	2022-05-18 11:24:12.587149+02
nc_93ugpir0y6x9re	vw_h61l96me3nfmwj	cl_j8lzgbvoagbw4d	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.606229+02	2022-05-18 11:24:12.606229+02
nc_6zjrjb5go4zq3o	vw_352nd6p9zohj5w	cl_fpfamco3hhd3sv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.619009+02	2022-05-18 11:24:12.619009+02
nc_yuz7byuh9p9w5s	vw_mauaaqk5b7v0rr	cl_eo23q73pocm3d5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.621744+02	2022-05-18 11:24:12.621744+02
nc_5vfci9dpvga0er	vw_dejssdomuxg3uw	cl_myyzrukjep7quc	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.624482+02	2022-05-18 11:24:12.624482+02
nc_rr22wfm7nea52w	vw_5w6mg5ux0ryjjl	cl_2sucn53aey97zz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.628504+02	2022-05-18 11:24:12.628504+02
nc_kzo8g8ue7ws2rs	vw_u2jzjq9smernol	cl_oorghne0ybyrmk	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.649463+02	2022-05-18 11:24:12.649463+02
nc_6twnvzl5d3n6kz	vw_avtdalhk6frgn9	cl_10h4tlr5w8w9l9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.652321+02	2022-05-18 11:24:12.652321+02
nc_04lzayn0h7lj9p	vw_bg9irqfa68gtl3	cl_6f5tq9drcdxwh7	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.659902+02	2022-05-18 11:24:12.659902+02
nc_rq1wkcpnabu6ma	vw_dejssdomuxg3uw	cl_ycfpzz491h9tjf	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.667578+02	2022-05-18 11:24:12.667578+02
nc_j2sgu54kp8pbmp	vw_wflzr4fus2ngby	cl_7uz6ir4vjgn1cx	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.683676+02	2022-05-18 11:24:12.683676+02
nc_0hue3o61ycpmqr	vw_etd91poe0pvyld	cl_pu8p0rjz0tzntw	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.686336+02	2022-05-18 11:24:12.686336+02
nc_1awgwi66lxnsy4	vw_dejssdomuxg3uw	cl_kd0zz03498jx5q	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.69622+02	2022-05-18 11:24:12.69622+02
nc_2asd14zcscmc8h	vw_srnlvrf0gn6rul	cl_agwmleh0fxpuyj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.739839+02	2022-05-18 11:24:12.739839+02
nc_b1voyibfq0wuu7	vw_t8o4jwxtrx9nf0	cl_2kj2q3s1cv02g7	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.742216+02	2022-05-18 11:24:12.742216+02
nc_cggzz95bzsovln	vw_avtdalhk6frgn9	cl_sw2ka2n7mjx46z	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:12.751803+02	2022-05-18 11:24:12.751803+02
nc_xs4p1yu93fd7nc	vw_ewjjjsi3mbc5h9	cl_waudmda08mbill	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.989026+02	2022-05-18 11:24:11.989026+02
nc_2xgr52wnlz00i4	vw_du7zrxhj51z27u	cl_cwask6ptfuopet	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.032865+02	2022-05-18 11:24:12.032865+02
nc_dvh12ran0qyzrv	vw_8kwir8e38t3fc9	cl_0ae4rqt28h623c	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.052939+02	2022-05-18 11:24:12.052939+02
nc_k9nk4njvowv7kp	vw_tsu33qayws91ma	cl_mglgw6q3hpyxw5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.059751+02	2022-05-18 11:24:12.059751+02
nc_e2wd7yq8ndnn20	vw_ewjjjsi3mbc5h9	cl_bkvskjvf9kp0sk	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.089296+02	2022-05-18 11:24:12.089296+02
nc_v6moz45cctqd3u	vw_dejssdomuxg3uw	cl_hrs9ijvoaus9ls	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.118216+02	2022-05-18 11:24:12.118216+02
nc_27896j6s5sfyls	vw_r15acq81tqlkxp	cl_f2d7hc4bji50xx	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.123792+02	2022-05-18 11:24:12.123792+02
nc_6fj0izbg4auw7n	vw_h61l96me3nfmwj	cl_wn3aloaldty24k	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.16308+02	2022-05-18 11:24:12.16308+02
nc_kuzka0lvvk9a2p	vw_7ppnzs3y6eip9i	cl_enjcwnqd9lm0hl	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.18692+02	2022-05-18 11:24:12.18692+02
nc_31l1kleq11xgev	vw_gl8cwxr6m3iq38	cl_swwufmsw23vux6	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.248278+02	2022-05-18 11:24:12.248278+02
nc_86e76dzqpglpp0	vw_lalt4t2bjas0gj	cl_8y5bfy3i2yupmi	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.253714+02	2022-05-18 11:24:12.253714+02
nc_p4bqi4i02en1nd	vw_6exoiii01n21je	cl_jg8un0z5v4mbq0	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.260537+02	2022-05-18 11:24:12.260537+02
nc_hbqz7j0dx5me0g	vw_ztmg51h4hma3hj	cl_nv1tein75ejh1j	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.283529+02	2022-05-18 11:24:12.283529+02
nc_3egngd2twiy0tv	vw_h61l96me3nfmwj	cl_wnu5t8ebtmy6o5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.313904+02	2022-05-18 11:24:12.313904+02
nc_v7f1w1nqlzkmc2	vw_gl8cwxr6m3iq38	cl_eb5dfr8wdn1kne	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.319384+02	2022-05-18 11:24:12.319384+02
nc_3xngru26qsf1ju	vw_lalt4t2bjas0gj	cl_4xfv5oniyd1uke	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.324224+02	2022-05-18 11:24:12.324224+02
nc_w7lbwizxvs7lvk	vw_u2jzjq9smernol	cl_imu0a0bfg69jjz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.3906+02	2022-05-18 11:24:12.3906+02
nc_lvlnxyu87eqsjx	vw_ewjjjsi3mbc5h9	cl_st7iuxws65atp5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.397958+02	2022-05-18 11:24:12.397958+02
nc_4u9ezzllbjcgp7	vw_6exoiii01n21je	cl_v0sietsica1ttd	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.410802+02	2022-05-18 11:24:12.410802+02
nc_7w8bsn38zwuty4	vw_tsu33qayws91ma	cl_8ux969grq7yqo6	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.435211+02	2022-05-18 11:24:12.435211+02
nc_0obrck8145fntp	vw_gl8cwxr6m3iq38	cl_sqj0wjnlu6z8j8	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.464836+02	2022-05-18 11:24:12.464836+02
nc_0m5rrh4ad6f4a1	vw_u2jzjq9smernol	cl_vfsxjg732tgz6a	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.50802+02	2022-05-18 11:24:12.50802+02
nc_y23wryvmk46px0	vw_andyomdmvq8tv6	cl_dd7bou38xmdibz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.514721+02	2022-05-18 11:24:12.514721+02
nc_mqyy93wzvr99lq	vw_352nd6p9zohj5w	cl_t4lk8w5lb5xsa6	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.520541+02	2022-05-18 11:24:12.520541+02
nc_es2nsy8ofio7pt	vw_etd91poe0pvyld	cl_8bww0dmnwaktnp	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.607116+02	2022-05-18 11:24:12.607116+02
nc_l2ts3wx1i6vklv	vw_andyomdmvq8tv6	cl_nj629z71fvb8xr	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.613673+02	2022-05-18 11:24:12.613673+02
nc_b7scwx3rjzdqlx	vw_t8o4jwxtrx9nf0	cl_rg5pf6ur5s8ofd	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.619613+02	2022-05-18 11:24:12.619613+02
nc_p1k5jycc3pr22l	vw_wflzr4fus2ngby	cl_69guec8gq6p2oe	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.6472+02	2022-05-18 11:24:12.6472+02
nc_z40hwx42l6vcbb	vw_h61l96me3nfmwj	cl_m1wh0xmljkxlzu	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.681458+02	2022-05-18 11:24:12.681458+02
nc_dmhqbexbszram8	vw_lalt4t2bjas0gj	cl_pnrrjhwfpyqi12	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.686996+02	2022-05-18 11:24:12.686996+02
nc_hgu35fp9ir09gp	vw_hk5zatfifol9yl	cl_eagzbs90wvmutu	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.710171+02	2022-05-18 11:24:12.710171+02
nc_hbtdxi6oc7lt6n	vw_t8o4jwxtrx9nf0	cl_ppkvlb1bhb29sv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:12.769134+02	2022-05-18 11:24:12.769134+02
nc_apf4rmj1a137w7	vw_xfg9ot5ntsq77q	cl_tbvbq12n5xrtso	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.93741+02	2022-05-18 11:24:12.93741+02
nc_u988mifdqbufeh	vw_352nd6p9zohj5w	cl_2ruapfr842h3bp	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.941718+02	2022-05-18 11:24:12.941718+02
nc_204do8y4474a1i	vw_7ppnzs3y6eip9i	cl_8cezepfnbj7t2k	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.9451+02	2022-05-18 11:24:12.9451+02
nc_vbyizbgpqvakoz	vw_etd91poe0pvyld	cl_j32rfmzk6jy8en	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:12.947918+02	2022-05-18 11:24:12.947918+02
nc_un0nbt7lswzhvw	vw_andyomdmvq8tv6	cl_ftb3421h4x8xxr	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:12.950228+02	2022-05-18 11:24:12.950228+02
nc_c9l1yjydbbq67b	vw_tsu33qayws91ma	cl_nbodzf77l55cru	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:13.029968+02	2022-05-18 11:24:13.029968+02
nc_z02drq7ze7f5ft	vw_q1cl9wv3hhuvca	cl_hsqf01u6x6uwz7	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:13.032059+02	2022-05-18 11:24:13.032059+02
nc_crfxad409tlxbj	vw_jbveqkpzj241da	cl_ikur0598w6hz3f	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:13.034557+02	2022-05-18 11:24:13.034557+02
nc_hvq94yi1rufk1k	vw_hk5zatfifol9yl	cl_q3y938j1f63exf	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:13.087146+02	2022-05-18 11:24:13.087146+02
nc_ztv3l10mo4ebfq	vw_u2jzjq9smernol	cl_ru32bw7buii6la	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	22	2022-05-18 11:24:13.113468+02	2022-05-18 11:24:13.113468+02
nc_n6t3ocek0m7zva	vw_jbveqkpzj241da	cl_291qx9ou1akw8y	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.990738+02	2022-05-18 11:24:11.990738+02
nc_gweuhlsci8cucj	vw_7fueohglirse69	cl_tgvxn298rrqb61	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.001911+02	2022-05-18 11:24:12.001911+02
nc_j4lsk35pxdbtdk	vw_51d0c7bmbclwb8	cl_t4dwkmgadehh7w	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.017743+02	2022-05-18 11:24:12.017743+02
nc_wmk0tvsdhyeq2n	vw_81imke2v3e15l5	cl_bbui77uct3fuc5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.047195+02	2022-05-18 11:24:12.047195+02
nc_jhozdwv10zcxm7	vw_etd91poe0pvyld	cl_da7co98ztmz4v6	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.088035+02	2022-05-18 11:24:12.088035+02
nc_p9ak10361ta5xo	vw_7ppnzs3y6eip9i	cl_7fuircxtsky4ng	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.105349+02	2022-05-18 11:24:12.105349+02
nc_n3odz4r18jnvyg	vw_idcyonon9t2t5l	cl_3eagrih1e0zl3n	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.131053+02	2022-05-18 11:24:12.131053+02
nc_qnfndkocwaf9yy	vw_hk5zatfifol9yl	cl_3gm2h79uxhdzcc	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.171176+02	2022-05-18 11:24:12.171176+02
nc_dcxzg5mtk7r2tb	vw_lalt4t2bjas0gj	cl_l08wx8t5fcd9lu	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.181318+02	2022-05-18 11:24:12.181318+02
nc_7de5uh4s4x567w	vw_6exoiii01n21je	cl_2b7kd2rxaa6i09	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.188989+02	2022-05-18 11:24:12.188989+02
nc_ss2tvt3qwaffed	vw_yrmq5jknxj9sx6	cl_0pok7tlgudc71r	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.214089+02	2022-05-18 11:24:12.214089+02
nc_7wb9d9hw8u99v3	vw_ewjjjsi3mbc5h9	cl_3kjqc5f1aywd4g	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.250065+02	2022-05-18 11:24:12.250065+02
nc_dpm4msne3mgab9	vw_mauaaqk5b7v0rr	cl_1qvn468mxp9xd0	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.264067+02	2022-05-18 11:24:12.264067+02
nc_izcx210yya0szk	vw_ewjjjsi3mbc5h9	cl_s38bav6sh7c4en	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.318382+02	2022-05-18 11:24:12.318382+02
nc_csgoj8m2hnt22r	vw_bg9irqfa68gtl3	cl_lgjbx0b7bhizkc	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.33433+02	2022-05-18 11:24:12.33433+02
nc_9ghbpo4basht5r	vw_srnlvrf0gn6rul	cl_3a45lb54s4im5e	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.354761+02	2022-05-18 11:24:12.354761+02
nc_hqboq7vfuxd1la	vw_jbveqkpzj241da	cl_zpybi9h1u394ao	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.388931+02	2022-05-18 11:24:12.388931+02
nc_18ynfudo3sed9m	vw_gl8cwxr6m3iq38	cl_l4uk1f2f7g5wxq	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.402453+02	2022-05-18 11:24:12.402453+02
nc_5w75p0cnns0b9m	vw_wflzr4fus2ngby	cl_tux3x9lo761ow6	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.456374+02	2022-05-18 11:24:12.456374+02
nc_mr2t45mwxnzj7g	vw_ewjjjsi3mbc5h9	cl_dri9nl4xksb0u6	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.463717+02	2022-05-18 11:24:12.463717+02
nc_pqf4ospn465wv8	vw_352nd6p9zohj5w	cl_ybvnp05fszzbqg	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.470736+02	2022-05-18 11:24:12.470736+02
nc_qu07u2brk9nsl3	vw_5w6mg5ux0ryjjl	cl_grx4o67kbsjtct	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.487102+02	2022-05-18 11:24:12.487102+02
nc_tfc9vgjaht79fu	vw_6exoiii01n21je	cl_dp2o9412pva5ph	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.521977+02	2022-05-18 11:24:12.521977+02
nc_9i3x7e2ywoesum	vw_bg9irqfa68gtl3	cl_sf0a0bdn3w6jfi	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.524733+02	2022-05-18 11:24:12.524733+02
nc_1fv2k3fqevrd3s	vw_dejssdomuxg3uw	cl_a00e6lm16vrvm9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.533446+02	2022-05-18 11:24:12.533446+02
nc_jkf27lf7jju14h	vw_u2jzjq9smernol	cl_rfdwwrsb228fls	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.563023+02	2022-05-18 11:24:12.563023+02
nc_vu487rmp2x25jp	vw_hk5zatfifol9yl	cl_0biojey4j3nxim	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.565466+02	2022-05-18 11:24:12.565466+02
nc_r7t504gltozik5	vw_wflzr4fus2ngby	cl_uqz0n2ikliepax	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.605651+02	2022-05-18 11:24:12.605651+02
nc_nfgymiyxhimisv	vw_pcdpz0lqr00s9k	cl_ygqq2ohbanufuo	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.608397+02	2022-05-18 11:24:12.608397+02
nc_c58id9tabu0r3z	vw_jq0zcol66s3voo	cl_llodox6wyofvr5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.615027+02	2022-05-18 11:24:12.615027+02
nc_9c5wcxhvn7yowg	vw_jbveqkpzj241da	cl_2pskm13svz0wnf	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.651834+02	2022-05-18 11:24:12.651834+02
nc_jsv8i6qagf480w	vw_6exoiii01n21je	cl_pnacxw8tomau6f	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.656412+02	2022-05-18 11:24:12.656412+02
nc_l1mqol1q2a1u7c	vw_u2jzjq9smernol	cl_md0irktw1j1uvk	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.683277+02	2022-05-18 11:24:12.683277+02
nc_394gzkcenxq16e	vw_pcdpz0lqr00s9k	cl_q1v7cutyjfyvov	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.685799+02	2022-05-18 11:24:12.685799+02
nc_9ukj09vt7def4t	vw_avtdalhk6frgn9	cl_6smvoxg4vqvehd	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.708652+02	2022-05-18 11:24:12.708652+02
nc_6rkrbibc9w39lm	vw_bg9irqfa68gtl3	cl_f41hgjkayi76ap	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.71865+02	2022-05-18 11:24:12.71865+02
nc_n0vpqg2jf8eajj	vw_u2jzjq9smernol	cl_w9aupuppuv6o2k	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.734879+02	2022-05-18 11:24:12.734879+02
nc_27xqiwv4godno7	vw_ea8mjggb90pnnt	cl_pc6ytykx2z82a6	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:12.753837+02	2022-05-18 11:24:12.753837+02
nc_73n6n1ccismkox	vw_ea8mjggb90pnnt	cl_sn5wo6n2vcf8tr	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:12.768091+02	2022-05-18 11:24:12.768091+02
nc_1i6ph0ptf1x77a	vw_ea8mjggb90pnnt	cl_rwie4in2uxuz1i	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	18	2022-05-18 11:24:12.77945+02	2022-05-18 11:24:12.77945+02
nc_04a9qil496khuh	vw_ea8mjggb90pnnt	cl_w0y5qxjh7p18ba	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	19	2022-05-18 11:24:12.785282+02	2022-05-18 11:24:12.785282+02
nc_mu6848q192mo04	vw_ea8mjggb90pnnt	cl_aoi05i0khwxm5g	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	20	2022-05-18 11:24:12.790219+02	2022-05-18 11:24:12.790219+02
nc_isqg77qmx4hmrz	vw_ea8mjggb90pnnt	cl_xej197pp067b4h	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	21	2022-05-18 11:24:12.79669+02	2022-05-18 11:24:12.79669+02
nc_ozjmu8q8z2aayg	vw_ea8mjggb90pnnt	cl_6iz63txm4hfh7f	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	22	2022-05-18 11:24:12.800422+02	2022-05-18 11:24:12.800422+02
nc_cabdwqyw0fj6pq	vw_idcyonon9t2t5l	cl_dyz994lxx7yesp	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.936684+02	2022-05-18 11:24:12.936684+02
nc_jljr6w3tb0pq5m	vw_dzgaxceud97gqt	cl_d8fg3jivyrrhux	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.938564+02	2022-05-18 11:24:12.938564+02
nc_thbh1mufc65xbp	vw_7fueohglirse69	cl_318j4uypmgfmdn	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.940606+02	2022-05-18 11:24:12.940606+02
nc_a9eftpyuoecl4w	vw_pcdpz0lqr00s9k	cl_p619rqzumc5ueb	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.943051+02	2022-05-18 11:24:12.943051+02
nc_yow2ar9isc5pok	vw_ewjjjsi3mbc5h9	cl_mg31nh0giwezx0	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	19	2022-05-18 11:24:12.946811+02	2022-05-18 11:24:12.946811+02
nc_w7irjq7piw1xkc	vw_pcdpz0lqr00s9k	cl_iajc7q7keo69iv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.991641+02	2022-05-18 11:24:11.991641+02
nc_dt2nfzsonvcm9e	vw_hk5zatfifol9yl	cl_e45d0ubf4cepmg	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.996249+02	2022-05-18 11:24:11.996249+02
nc_dqlx6h2lzgfyli	vw_n362p00huzpsod	cl_iwu0vwdgt2ax7s	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.002936+02	2022-05-18 11:24:12.002936+02
nc_nw7ib4btn4uniw	vw_ea8mjggb90pnnt	cl_obsjzit9bzkszx	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.005944+02	2022-05-18 11:24:12.005944+02
nc_ldvsii7mdhs1j2	vw_7ppnzs3y6eip9i	cl_9wmjbqjyqwqjfv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.014869+02	2022-05-18 11:24:12.014869+02
nc_lz8feuutw7zmbw	vw_q1cl9wv3hhuvca	cl_ttu9qrvvb56wps	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.018916+02	2022-05-18 11:24:12.018916+02
nc_jrixermkb23849	vw_idcyonon9t2t5l	cl_oq9xmxjup9n57n	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.043131+02	2022-05-18 11:24:12.043131+02
nc_5ruouzgguk091a	vw_dzgaxceud97gqt	cl_lbtwbffy5ca4x5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.082045+02	2022-05-18 11:24:12.082045+02
nc_z9mixduxkulux4	vw_sxajywo4jx9ml5	cl_5gkj8u9n2irl5t	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.086326+02	2022-05-18 11:24:12.086326+02
nc_2qtxy12h2ee3qb	vw_n362p00huzpsod	cl_hvpo5wfg6q9bhy	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.095599+02	2022-05-18 11:24:12.095599+02
nc_2c72puezwrg1dg	vw_lalt4t2bjas0gj	cl_xj1mgb1h75y03n	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.099394+02	2022-05-18 11:24:12.099394+02
nc_ei9lvy74mj7p5j	vw_go1u33h41idruk	cl_2va4hiqfh4srjj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.103251+02	2022-05-18 11:24:12.103251+02
nc_bzsgem8r1i3cz4	vw_mauaaqk5b7v0rr	cl_onax81ltzbvj6l	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.106439+02	2022-05-18 11:24:12.106439+02
nc_ldvhrtu3nvr788	vw_dzgaxceud97gqt	cl_ytfwqrsq3akllw	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.166388+02	2022-05-18 11:24:12.166388+02
nc_ox340xbfhluu2w	vw_sxajywo4jx9ml5	cl_luguf02s068frs	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.169375+02	2022-05-18 11:24:12.169375+02
nc_6fqe6gos8o85wp	vw_pcdpz0lqr00s9k	cl_8vqk98wddq5tto	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.17195+02	2022-05-18 11:24:12.17195+02
nc_evp7ojp1q3baq5	vw_ewjjjsi3mbc5h9	cl_sf4me3yekzw8kw	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.174561+02	2022-05-18 11:24:12.174561+02
nc_xuovst49xnrqgz	vw_andyomdmvq8tv6	cl_e0qvng7dplmklv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.182389+02	2022-05-18 11:24:12.182389+02
nc_t6yo7d6wbbbq2s	vw_q1cl9wv3hhuvca	cl_yf33uxsgmmtxzo	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.190237+02	2022-05-18 11:24:12.190237+02
nc_2xd385k56dpllw	vw_dzgaxceud97gqt	cl_i2d4othvsshdrf	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.240483+02	2022-05-18 11:24:12.240483+02
nc_w2ngb8qnyhl3b9	vw_sxajywo4jx9ml5	cl_a2e7sflk8yq825	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.248045+02	2022-05-18 11:24:12.248045+02
nc_3n6cht173d8gsj	vw_2lgkfm9g3xygp1	cl_9j9i6wt5dojluu	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.250775+02	2022-05-18 11:24:12.250775+02
nc_2ratdoh27qzywl	vw_jq0zcol66s3voo	cl_nrckcor6haqx4w	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.253386+02	2022-05-18 11:24:12.253386+02
nc_jjwpewdgrji3zn	vw_7fueohglirse69	cl_db2hgifuodn7cs	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.256298+02	2022-05-18 11:24:12.256298+02
nc_81kzd7d3k91mra	vw_go1u33h41idruk	cl_2l9fbdjfosuadr	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.259959+02	2022-05-18 11:24:12.259959+02
nc_fac2mj6jsqtpfd	vw_bg9irqfa68gtl3	cl_6knh4wr0379knt	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.268948+02	2022-05-18 11:24:12.268948+02
nc_xnvotg2tfzwwdq	vw_jbveqkpzj241da	cl_t92sq868h7db4w	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.312965+02	2022-05-18 11:24:12.312965+02
nc_vo612t1lu7maok	vw_sxajywo4jx9ml5	cl_eid5i5cfqi101t	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.316489+02	2022-05-18 11:24:12.316489+02
nc_3oidvxbcxa1qq5	vw_2lgkfm9g3xygp1	cl_wyit1f1uhnr4ux	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.319124+02	2022-05-18 11:24:12.319124+02
nc_9393ealittcxnk	vw_7fueohglirse69	cl_o4gbtsb3rs0i4x	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.324031+02	2022-05-18 11:24:12.324031+02
nc_tt7jn96w4igkng	vw_dejssdomuxg3uw	cl_elxzenlpnoman8	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.339613+02	2022-05-18 11:24:12.339613+02
nc_sbd9i0anjrx07d	vw_andyomdmvq8tv6	cl_yaupebwbnv4qxt	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.400734+02	2022-05-18 11:24:12.400734+02
nc_s9c5c5lch1o0at	vw_ea8mjggb90pnnt	cl_qzvzbr3q5bxe4f	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.40605+02	2022-05-18 11:24:12.40605+02
nc_k06pvc7aby0bww	vw_sxajywo4jx9ml5	cl_r0os9n1us2h7be	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.4542+02	2022-05-18 11:24:12.4542+02
nc_wxdfyot66nswxj	vw_h61l96me3nfmwj	cl_z7tkdlhl3cxjj9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.457523+02	2022-05-18 11:24:12.457523+02
nc_9zk5psdadf7mtv	vw_7ppnzs3y6eip9i	cl_hn4ew82cx6hg2s	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.467755+02	2022-05-18 11:24:12.467755+02
nc_2569pnfpiz77mt	vw_ea8mjggb90pnnt	cl_5qy65ad5mlzglb	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.471696+02	2022-05-18 11:24:12.471696+02
nc_5q3jjxdp8al1uv	vw_mauaaqk5b7v0rr	cl_8cf2j5ptfccdk1	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.474471+02	2022-05-18 11:24:12.474471+02
nc_9fnl0hpnrngrsg	vw_tsu33qayws91ma	cl_fvanz3lcymlorp	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.490968+02	2022-05-18 11:24:12.490968+02
nc_2m4hynnzgnl0mu	vw_jbveqkpzj241da	cl_t3yqx8owhtr4p7	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.507596+02	2022-05-18 11:24:12.507596+02
nc_2iu99vcylc4nri	vw_lalt4t2bjas0gj	cl_38ocyrkyz7o69r	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.520276+02	2022-05-18 11:24:12.520276+02
nc_s724l8zbemvuys	vw_ztmg51h4hma3hj	cl_xbrmwc3imzej7e	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.534202+02	2022-05-18 11:24:12.534202+02
nc_4ftnx577b0gxwv	vw_srnlvrf0gn6rul	cl_x7ssnak28x8548	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.536951+02	2022-05-18 11:24:12.536951+02
nc_scrip0dr6kdi3a	vw_5w6mg5ux0ryjjl	cl_rw8jc866mp0w5o	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.54028+02	2022-05-18 11:24:12.54028+02
nc_azuqsfbftwf0uz	vw_go1u33h41idruk	cl_ja4j53raurxvz1	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.571245+02	2022-05-18 11:24:12.571245+02
nc_jbd0t4ywa4jhcj	vw_bg9irqfa68gtl3	cl_amisrr6l4xmpft	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.573909+02	2022-05-18 11:24:12.573909+02
nc_tp96ne4g7l87gy	vw_srnlvrf0gn6rul	cl_h4vfb2b5flhu5c	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.628992+02	2022-05-18 11:24:12.628992+02
nc_k9ravypshnnkjt	vw_pcdpz0lqr00s9k	cl_d60otwkbbnwcl4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.649874+02	2022-05-18 11:24:12.649874+02
nc_xqpayn5b1fm0ha	vw_lalt4t2bjas0gj	cl_lxa3pmvssaszo6	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.654726+02	2022-05-18 11:24:12.654726+02
nc_y93at9ds4s5z5q	vw_5w6mg5ux0ryjjl	cl_gtql7p62i2525w	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.667776+02	2022-05-18 11:24:12.667776+02
nc_fu6emoztirepjv	vw_ea8mjggb90pnnt	cl_1cgi8am13r1fao	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.691523+02	2022-05-18 11:24:12.691523+02
nc_nete9jflw4a4qg	vw_srnlvrf0gn6rul	cl_rqvi328ikm9se5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.69656+02	2022-05-18 11:24:12.69656+02
nc_e4k4s8rnmmiwnb	vw_2lgkfm9g3xygp1	cl_ds7fd5osbe41sp	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.993564+02	2022-05-18 11:24:11.993564+02
nc_yzwi4tywi6w5f9	vw_wflzr4fus2ngby	cl_r39qpsifh4dd16	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:11.998395+02	2022-05-18 11:24:11.998395+02
nc_j4mmbzitw2ipvv	vw_andyomdmvq8tv6	cl_1jvx4978ntayw5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.004571+02	2022-05-18 11:24:12.004571+02
nc_rtapdgucz02mu0	vw_r15acq81tqlkxp	cl_s7mtg7jpphuwfw	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.038248+02	2022-05-18 11:24:12.038248+02
nc_tnpv4b1q05l0k1	vw_u2jzjq9smernol	cl_dan2jfbs2nmq5u	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.080506+02	2022-05-18 11:24:12.080506+02
nc_sqz3dqrfdz5nv7	vw_ipcsobcohj1b6c	cl_cxqf2kqwmat0k3	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.084436+02	2022-05-18 11:24:12.084436+02
nc_r0gdtgzhiz1ls0	vw_7fueohglirse69	cl_u4mw1fxp0si97k	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.093207+02	2022-05-18 11:24:12.093207+02
nc_oa2up243v4yw01	vw_q1cl9wv3hhuvca	cl_fwqr5zoahf4y1d	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.108151+02	2022-05-18 11:24:12.108151+02
nc_lkvbk4wk3fnhy4	vw_brn7rul3d130ex	cl_da3hlyk05rx1wf	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.112076+02	2022-05-18 11:24:12.112076+02
nc_jm1i6nodaqdmdd	vw_ipcsobcohj1b6c	cl_p0r4ptz7y81v5n	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.167966+02	2022-05-18 11:24:12.167966+02
nc_nwho96vsowxnc4	vw_avtdalhk6frgn9	cl_m6bwhgghyxz32g	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.181023+02	2022-05-18 11:24:12.181023+02
nc_pr36e985h0y5nw	vw_go1u33h41idruk	cl_0cp6rgcqobi280	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.185184+02	2022-05-18 11:24:12.185184+02
nc_etayntmhlpdwt6	vw_51d0c7bmbclwb8	cl_j56fpioga326or	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.188669+02	2022-05-18 11:24:12.188669+02
nc_8ovwbxxilf3ady	vw_hk5zatfifol9yl	cl_rfzkjum97sm4px	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.246983+02	2022-05-18 11:24:12.246983+02
nc_3yajs6shmpyrfq	vw_7ppnzs3y6eip9i	cl_pi3yol2gqnb28k	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.258316+02	2022-05-18 11:24:12.258316+02
nc_vm1eh5ckyl3nlh	vw_lr57pe50bbmjar	cl_s04j0dacnnx6jh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.267174+02	2022-05-18 11:24:12.267174+02
nc_by76o736qi57kw	vw_u2jzjq9smernol	cl_t9bju32ijpe5ei	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.315274+02	2022-05-18 11:24:12.315274+02
nc_0w8qzsad3icjmw	vw_hk5zatfifol9yl	cl_v63cf04xvz97pb	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.317908+02	2022-05-18 11:24:12.317908+02
nc_hqhjlm6ebtz05x	vw_ea8mjggb90pnnt	cl_e72jcq87j4l87q	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.325377+02	2022-05-18 11:24:12.325377+02
nc_1axdlx2u5181yb	vw_h61l96me3nfmwj	cl_p8rug1g5fmx2rz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.391569+02	2022-05-18 11:24:12.391569+02
nc_gd3b2qpmnnk2cd	vw_2lgkfm9g3xygp1	cl_6c3qkavaz14cfj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.396261+02	2022-05-18 11:24:12.396261+02
nc_ql0lrf7qdgz12k	vw_lalt4t2bjas0gj	cl_v0nkthutghwjgf	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.404987+02	2022-05-18 11:24:12.404987+02
nc_lg05h87tfli95k	vw_352nd6p9zohj5w	cl_jcmfany6r3buss	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.407841+02	2022-05-18 11:24:12.407841+02
nc_31q8b9cfr1ca6k	vw_ztmg51h4hma3hj	cl_f3lqniknfi47bd	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.424028+02	2022-05-18 11:24:12.424028+02
nc_8mnmjsa381m8dr	vw_srnlvrf0gn6rul	cl_m6xqxi37t6xbx8	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.429229+02	2022-05-18 11:24:12.429229+02
nc_tawm6t742x7iqv	vw_pcdpz0lqr00s9k	cl_1ddmdm9b8op8jw	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.456013+02	2022-05-18 11:24:12.456013+02
nc_tkm4i838amtjxu	vw_hk5zatfifol9yl	cl_oxsn7xt4obc7r2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.459083+02	2022-05-18 11:24:12.459083+02
nc_qh2wxv97ajf65r	vw_andyomdmvq8tv6	cl_x1knxeqizscbih	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.46338+02	2022-05-18 11:24:12.46338+02
nc_tuscdjioo43gxt	vw_lalt4t2bjas0gj	cl_lxhi365f4uspl0	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.470443+02	2022-05-18 11:24:12.470443+02
nc_k6vap39y6w2hhh	vw_srnlvrf0gn6rul	cl_uwmjhs1h96b5j7	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.486925+02	2022-05-18 11:24:12.486925+02
nc_qjvlk8u4kwm3vf	vw_h61l96me3nfmwj	cl_qy73tlix63j4b7	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.509729+02	2022-05-18 11:24:12.509729+02
nc_rw95vyxqrsd8of	vw_7ppnzs3y6eip9i	cl_bmdab93tjhr67q	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.518895+02	2022-05-18 11:24:12.518895+02
nc_1p5kmwzn6euf74	vw_jbveqkpzj241da	cl_xrpbnrlguot0p4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.565265+02	2022-05-18 11:24:12.565265+02
nc_k2yqitxnuf5srw	vw_andyomdmvq8tv6	cl_06o2ocohphf0sv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.568218+02	2022-05-18 11:24:12.568218+02
nc_qv8j66iulmz6c7	vw_ewjjjsi3mbc5h9	cl_fl7eap9v6vs6k5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.570399+02	2022-05-18 11:24:12.570399+02
nc_vp4wpoyai2ldd1	vw_mauaaqk5b7v0rr	cl_st7kgq3v75o7np	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.575744+02	2022-05-18 11:24:12.575744+02
nc_swki61703wl5rb	vw_dejssdomuxg3uw	cl_rvh480evb8owh2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.580281+02	2022-05-18 11:24:12.580281+02
nc_lwtwvgrze6i77s	vw_go1u33h41idruk	cl_qsksqgb6y745zn	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.614758+02	2022-05-18 11:24:12.614758+02
nc_zm7gp21a8s6ctp	vw_bg9irqfa68gtl3	cl_2nyio17f5epmq0	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.618172+02	2022-05-18 11:24:12.618172+02
nc_08skgpqaqophxz	vw_n362p00huzpsod	cl_u1du84u11xn4q7	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.653888+02	2022-05-18 11:24:12.653888+02
nc_9nlyw6xfq8kt2w	vw_ea8mjggb90pnnt	cl_5x63ldetbc36fe	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.663765+02	2022-05-18 11:24:12.663765+02
nc_z0d87lhdtssdrs	vw_hk5zatfifol9yl	cl_7ch3641kd324n7	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.682873+02	2022-05-18 11:24:12.682873+02
nc_uhaf2xnk5phu6x	vw_h61l96me3nfmwj	cl_wy8brvdnpgy3d3	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.708413+02	2022-05-18 11:24:12.708413+02
nc_8wxvdtyry1x8ql	vw_t8o4jwxtrx9nf0	cl_gvn88mg3ykq5n8	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.722772+02	2022-05-18 11:24:12.722772+02
nc_53itt7p6fkbnvt	vw_etd91poe0pvyld	cl_p2vlwzdwz6g4t3	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.734567+02	2022-05-18 11:24:12.734567+02
nc_54raju4wyq63no	vw_lalt4t2bjas0gj	cl_np4oitaanopvi0	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:12.755621+02	2022-05-18 11:24:12.755621+02
nc_osdkoni7u7f0os	vw_lalt4t2bjas0gj	cl_bvt77xombx4esv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:12.769756+02	2022-05-18 11:24:12.769756+02
nc_4w9358f5rvfel4	vw_lalt4t2bjas0gj	cl_2jdqr32vfn4gw4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	18	2022-05-18 11:24:12.780849+02	2022-05-18 11:24:12.780849+02
nc_5ttndznvkxx6pe	vw_lalt4t2bjas0gj	cl_ddia0w2o8xb2uv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	19	2022-05-18 11:24:12.786781+02	2022-05-18 11:24:12.786781+02
nc_dt9fgdw4pvby71	vw_lalt4t2bjas0gj	cl_xh3pswy5ur5ov9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	20	2022-05-18 11:24:12.791937+02	2022-05-18 11:24:12.791937+02
nc_uy83vv07h5y2s3	vw_81imke2v3e15l5	cl_t3dlbbwzew7sam	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.936826+02	2022-05-18 11:24:12.936826+02
nc_xc1wwc1exqbxa2	vw_go1u33h41idruk	cl_6xpxnl95np7q1z	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.006377+02	2022-05-18 11:24:12.006377+02
nc_agei95b5xaowwg	vw_yrmq5jknxj9sx6	cl_m1g535oorspd4b	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.040034+02	2022-05-18 11:24:12.040034+02
nc_zqaz9yrupky1nz	vw_h61l96me3nfmwj	cl_x9rp3wek2sqmkr	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.076718+02	2022-05-18 11:24:12.076718+02
nc_6r5uqan0yiijeq	vw_gl8cwxr6m3iq38	cl_xdp0ttgvijpllq	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.082353+02	2022-05-18 11:24:12.082353+02
nc_1ruq1rulweyyoa	vw_gl8cwxr6m3iq38	cl_vhim94i6ia8ld4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.16664+02	2022-05-18 11:24:12.16664+02
nc_k61b5cwgxvonch	vw_352nd6p9zohj5w	cl_hffjtbw2nw06p1	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.187237+02	2022-05-18 11:24:12.187237+02
nc_3cekgjtgtxtv8p	vw_r15acq81tqlkxp	cl_mtad5g6leynhzh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.207408+02	2022-05-18 11:24:12.207408+02
nc_kmyptsu8y21xen	vw_brn7rul3d130ex	cl_k4waejjah5twfj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.265565+02	2022-05-18 11:24:12.265565+02
nc_kv12jyjq9aeemx	vw_dzgaxceud97gqt	cl_hntenvlef6j2xi	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.308384+02	2022-05-18 11:24:12.308384+02
nc_u0ex5zg9bdou99	vw_hk5zatfifol9yl	cl_m5ihdmbagtnr1m	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.393854+02	2022-05-18 11:24:12.393854+02
nc_j8wxlty8mxguh0	vw_bg9irqfa68gtl3	cl_biry9fx9h2mhne	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.409884+02	2022-05-18 11:24:12.409884+02
nc_ylj2c5azijvhwy	vw_dejssdomuxg3uw	cl_cvotd774eyvz5m	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.415325+02	2022-05-18 11:24:12.415325+02
nc_2szjcojju4whcg	vw_bg9irqfa68gtl3	cl_d8isivipdmp7ka	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.472101+02	2022-05-18 11:24:12.472101+02
nc_wsneqvei8lniku	vw_dejssdomuxg3uw	cl_mcx1wq0b8d1q8n	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.479536+02	2022-05-18 11:24:12.479536+02
nc_7a14jp5v8na9nk	vw_hk5zatfifol9yl	cl_m2mfi2laj0zcuv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.511678+02	2022-05-18 11:24:12.511678+02
nc_k93srg5rn5p5os	vw_h61l96me3nfmwj	cl_qcdyzyflhbcxam	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.56024+02	2022-05-18 11:24:12.56024+02
nc_6qrjfdi4tlq8p2	vw_jq0zcol66s3voo	cl_o5umya2z9dueqy	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.567087+02	2022-05-18 11:24:12.567087+02
nc_j4ezxqurivdihr	vw_tsu33qayws91ma	cl_1821goipxujpwj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.590411+02	2022-05-18 11:24:12.590411+02
nc_xpla2wxnby0xw2	vw_n362p00huzpsod	cl_crrp1b4xmw6rye	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.60986+02	2022-05-18 11:24:12.60986+02
nc_8lmr60s52380yl	vw_ewjjjsi3mbc5h9	cl_rzvj3urxhsfmli	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.616783+02	2022-05-18 11:24:12.616783+02
nc_956s2wzkn8saxx	vw_mauaaqk5b7v0rr	cl_hkd55j8p8gdukv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.661446+02	2022-05-18 11:24:12.661446+02
nc_uxyf5khrg74a2r	vw_srnlvrf0gn6rul	cl_y2q5dfwbxx1n58	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.66803+02	2022-05-18 11:24:12.66803+02
nc_6wq7ykffmrecsf	vw_andyomdmvq8tv6	cl_fpfcswiqdsstff	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.68934+02	2022-05-18 11:24:12.68934+02
nc_sv71yflh9vrf3m	vw_pcdpz0lqr00s9k	cl_nacdeb0qxxghy1	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.714462+02	2022-05-18 11:24:12.714462+02
nc_ecung0rpkoo9e8	vw_ewjjjsi3mbc5h9	cl_xty7soa1r3z01z	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.719774+02	2022-05-18 11:24:12.719774+02
nc_uvu8urcd0loaci	vw_avtdalhk6frgn9	cl_ufirlt09c46y33	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.733206+02	2022-05-18 11:24:12.733206+02
nc_0bnt3ibctyk4te	vw_lalt4t2bjas0gj	cl_cou9zyhi047rfz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.737838+02	2022-05-18 11:24:12.737838+02
nc_xv1s3ld3myi1vd	vw_andyomdmvq8tv6	cl_n7bvxl6hrnzucd	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:12.766658+02	2022-05-18 11:24:12.766658+02
nc_f31kmshzmwl9vd	vw_c525afyrdh8nqn	cl_9w6yddj9cco5ke	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.937719+02	2022-05-18 11:24:12.937719+02
nc_x5srd8pbanpnin	vw_gl8cwxr6m3iq38	cl_ltn26rjondvorz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.93975+02	2022-05-18 11:24:12.93975+02
nc_9mx0gb5p96b6s5	vw_lr57pe50bbmjar	cl_k8gnkl4qn9ylb2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.941925+02	2022-05-18 11:24:12.941925+02
nc_i32346xxbghg3w	vw_5w6mg5ux0ryjjl	cl_6skr9pfy1fpbdh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.945568+02	2022-05-18 11:24:12.945568+02
nc_8wo7eq0exlkhsh	vw_jq0zcol66s3voo	cl_n13v5fxz7h0dgn	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:13.032391+02	2022-05-18 11:24:13.032391+02
nc_qvong2ytu1slgo	vw_u2jzjq9smernol	cl_uqphjc9xpcdo96	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	20	2022-05-18 11:24:13.034861+02	2022-05-18 11:24:13.034861+02
nc_t5bthbrls6x6ps	vw_ewjjjsi3mbc5h9	cl_qdrhybis5cbz4b	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	20	2022-05-18 11:24:13.037329+02	2022-05-18 11:24:13.037329+02
nc_er0zsgnxj5tng5	vw_h61l96me3nfmwj	cl_8cvwgubyxtqrep	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	35	2022-05-18 11:24:13.039593+02	2022-05-18 11:24:13.039593+02
nc_6tjzq0wvtkvvx8	vw_jq0zcol66s3voo	cl_il02qiiyaybb14	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:13.087474+02	2022-05-18 11:24:13.087474+02
nc_nyyfgma63ykjuk	vw_t8o4jwxtrx9nf0	cl_kwcxljamy9qed3	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	19	2022-05-18 11:24:13.089574+02	2022-05-18 11:24:13.089574+02
nc_ujibtki8ayhc4o	vw_srnlvrf0gn6rul	cl_fpxyzzfs6zlpdw	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	19	2022-05-18 11:24:13.123399+02	2022-05-18 11:24:13.123399+02
nc_bla0maov67miqi	vw_srnlvrf0gn6rul	cl_jtqvujcd57ffc1	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	20	2022-05-18 11:24:13.143242+02	2022-05-18 11:24:13.143242+02
nc_xuyxqwkri8sv7u	vw_srnlvrf0gn6rul	cl_7yabknvt1yvb2p	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	21	2022-05-18 11:24:13.162839+02	2022-05-18 11:24:13.162839+02
nc_ti0mhtcsg4vd3m	vw_srnlvrf0gn6rul	cl_87tqdqa6v32q7t	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	22	2022-05-18 11:24:13.170992+02	2022-05-18 11:24:13.170992+02
nc_me6xv7vbqw8xsi	vw_srnlvrf0gn6rul	cl_vbpxe8p4sakrfa	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	23	2022-05-18 11:24:13.178344+02	2022-05-18 11:24:13.178344+02
nc_4b13cfivz0a06u	vw_srnlvrf0gn6rul	cl_zc7rvd1ff56cib	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	24	2022-05-18 11:24:13.185168+02	2022-05-18 11:24:13.185168+02
nc_htx2npwy2ep73v	vw_29tn3xki80n7e3	cl_2n98v3z6z9y6eo	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	3	2022-05-18 11:24:12.121039+02	2022-05-18 11:56:38.998077+02
nc_44xmky7u99p8ah	vw_29tn3xki80n7e3	cl_2hvbzdhayp3m4d	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	9	2022-05-18 11:24:12.537291+02	2022-05-18 11:56:38.998077+02
nc_e1jvfozhu6nw39	vw_29tn3xki80n7e3	cl_nzd6ey7esqq626	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	14	2022-05-18 11:24:12.724021+02	2022-05-18 11:56:38.998077+02
nc_deh68a2ks63l40	vw_6exoiii01n21je	cl_37jdt2wdmn2fsh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.00799+02	2022-05-18 11:24:12.00799+02
nc_0mp4muwfz6cocd	vw_dejssdomuxg3uw	cl_rbx0v8hbfg9saz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	2	2022-05-18 11:24:12.023935+02	2022-05-18 11:24:12.023935+02
nc_ujpwe29wj7qin0	vw_jq0zcol66s3voo	cl_8xhockmpbsi2md	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.087537+02	2022-05-18 11:24:12.087537+02
nc_6qw447amcus7r5	vw_2lgkfm9g3xygp1	cl_fw3v3fz91x4v6e	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.090269+02	2022-05-18 11:24:12.090269+02
nc_m40g6ugdzw0c1p	vw_ztmg51h4hma3hj	cl_pwh6s0lomxswl7	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.130183+02	2022-05-18 11:24:12.130183+02
nc_k41cid58sd5iiy	vw_5w6mg5ux0ryjjl	cl_ep1ppfh31d8q0b	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.13498+02	2022-05-18 11:24:12.13498+02
nc_h6uege98ch9lw2	vw_tsu33qayws91ma	cl_hi6bftt57005lu	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.150696+02	2022-05-18 11:24:12.150696+02
nc_awver3qwk25j17	vw_u2jzjq9smernol	cl_4t4w33da2a7h6i	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.167755+02	2022-05-18 11:24:12.167755+02
nc_1fwcrwbuym84ww	vw_etd91poe0pvyld	cl_ll4pwrwaq6kj4i	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.170588+02	2022-05-18 11:24:12.170588+02
nc_2e1lkk01zpqtc3	vw_7fueohglirse69	cl_6a13znk1jszq23	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.175777+02	2022-05-18 11:24:12.175777+02
nc_k8vd1vv6ewjbgz	vw_srnlvrf0gn6rul	cl_2l1uwu50cen5px	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.208349+02	2022-05-18 11:24:12.208349+02
nc_ir0a7ida5fgrj4	vw_ztmg51h4hma3hj	cl_rk5bzouixq4oyt	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:12.213035+02	2022-05-18 11:24:12.213035+02
nc_rgd2cn7ueg1yfs	vw_andyomdmvq8tv6	cl_cdefrh9aw9kurm	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.254806+02	2022-05-18 11:24:12.254806+02
nc_wdcs4kj6e1j416	vw_352nd6p9zohj5w	cl_yws811j3krz71y	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.263302+02	2022-05-18 11:24:12.263302+02
nc_yp6l4oild2xjam	vw_t8o4jwxtrx9nf0	cl_8rjar6mb3mbmne	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.266752+02	2022-05-18 11:24:12.266752+02
nc_3c6uovwz8mox59	vw_andyomdmvq8tv6	cl_blr2fs7juvjvln	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.322817+02	2022-05-18 11:24:12.322817+02
nc_onhwgl93izkd8a	vw_7ppnzs3y6eip9i	cl_0gxxah1hq9herr	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.325169+02	2022-05-18 11:24:12.325169+02
nc_duhlzig3iel081	vw_mauaaqk5b7v0rr	cl_e98v7uxkxggrni	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:12.33327+02	2022-05-18 11:24:12.33327+02
nc_tkubv4dpwg13m0	vw_sxajywo4jx9ml5	cl_atylefeb8f6yfh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.391372+02	2022-05-18 11:24:12.391372+02
nc_siavgsp97236ob	vw_jq0zcol66s3voo	cl_xx7o3hptppb3jb	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.398862+02	2022-05-18 11:24:12.398862+02
nc_uqdo9yebnsakeh	vw_avtdalhk6frgn9	cl_5piiaqr116azp8	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.401966+02	2022-05-18 11:24:12.401966+02
nc_ukjkg7pp2bryrc	vw_t8o4jwxtrx9nf0	cl_adhc5aoe5s2dkl	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:12.416913+02	2022-05-18 11:24:12.416913+02
nc_i45mhegvtn7yw2	vw_u2jzjq9smernol	cl_c6hsuzyhqhcejg	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.455639+02	2022-05-18 11:24:12.455639+02
nc_08ok7afj7smfyf	vw_ztmg51h4hma3hj	cl_ewcsxrn7zdfg6x	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.483252+02	2022-05-18 11:24:12.483252+02
nc_lhwmzthznfuyw0	vw_jq0zcol66s3voo	cl_tox2sh660xnbo4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.515801+02	2022-05-18 11:24:12.515801+02
nc_wp1dc6pq3ugy33	vw_avtdalhk6frgn9	cl_0mczoi3l3e86wz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.518604+02	2022-05-18 11:24:12.518604+02
nc_bstn643rdbqlke	vw_ea8mjggb90pnnt	cl_tzcg2k5pfnkv2c	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.524235+02	2022-05-18 11:24:12.524235+02
nc_drf6ytw39h8z44	vw_mauaaqk5b7v0rr	cl_c18gv46qxxr0nd	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.527748+02	2022-05-18 11:24:12.527748+02
nc_6kypi7l0agnrn1	vw_n362p00huzpsod	cl_hpktuwr8acmf6w	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.568043+02	2022-05-18 11:24:12.568043+02
nc_6xbbo8f18gg0po	vw_352nd6p9zohj5w	cl_olwudcee6vaydm	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:12.575295+02	2022-05-18 11:24:12.575295+02
nc_c0yfojnl9j2hkk	vw_u2jzjq9smernol	cl_9yvg1tci6tf6xg	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.605104+02	2022-05-18 11:24:12.605104+02
nc_az5j2iffvoi87g	vw_hk5zatfifol9yl	cl_al38jy5qx7kc9t	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.607883+02	2022-05-18 11:24:12.607883+02
nc_fse7kdibxzxf9d	vw_avtdalhk6frgn9	cl_z86x61x8bf46a2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.611435+02	2022-05-18 11:24:12.611435+02
nc_bu5u0sw75cphik	vw_6exoiii01n21je	cl_bavibzttofdtq2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:12.617776+02	2022-05-18 11:24:12.617776+02
nc_jl7dhl91vkx1rc	vw_etd91poe0pvyld	cl_qvavs3z81dn4d8	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.651313+02	2022-05-18 11:24:12.651313+02
nc_n6ex9hgrbx5mfl	vw_andyomdmvq8tv6	cl_p83aw566opsxqw	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.658444+02	2022-05-18 11:24:12.658444+02
nc_o0wwgs4dq5w79v	vw_t8o4jwxtrx9nf0	cl_1ct8sqavrsm463	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.693079+02	2022-05-18 11:24:12.693079+02
nc_3kr1083tawsgfi	vw_lalt4t2bjas0gj	cl_122dzvsqblwsqd	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.718025+02	2022-05-18 11:24:12.718025+02
nc_63v8ko1j19tpz4	vw_andyomdmvq8tv6	cl_ukn6ghqw1xr48j	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:12.720632+02	2022-05-18 11:24:12.720632+02
nc_x3sal1sn8og2hd	vw_ewjjjsi3mbc5h9	cl_lw8tfi7qc221ob	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.741257+02	2022-05-18 11:24:12.741257+02
nc_4cf4fh02ax53bj	vw_etd91poe0pvyld	cl_66atpjp42alzgd	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:12.750808+02	2022-05-18 11:24:12.750808+02
nc_tupotxic1l0n5f	vw_jbveqkpzj241da	cl_d0if632wlv7hiu	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.942508+02	2022-05-18 11:24:12.942508+02
nc_f74m9sr4733ury	vw_n362p00huzpsod	cl_38ziv4dkkdh8iw	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:13.030646+02	2022-05-18 11:24:13.030646+02
nc_5ezi4u2j1zowjv	vw_h61l96me3nfmwj	cl_18yvnok8ybldha	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	36	2022-05-18 11:24:13.08787+02	2022-05-18 11:24:13.08787+02
nc_rwayvguv7ope0j	vw_29tn3xki80n7e3	cl_m4ojzvyzshvizr	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	16	2022-05-18 11:24:13.035284+02	2022-05-18 11:56:38.998077+02
nc_yxo3kx036bnic9	vw_h61l96me3nfmwj	cl_o6znciqwr4iuto	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.732991+02	2022-05-18 11:24:12.732991+02
nc_ifolbt3zg2i0ju	vw_ea8mjggb90pnnt	cl_9bpgd7aysn8o4k	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	15	2022-05-18 11:24:12.737567+02	2022-05-18 11:24:12.737567+02
nc_6i0yi47mvza6ta	vw_h61l96me3nfmwj	cl_qedxiu1jd91xje	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:12.749414+02	2022-05-18 11:24:12.749414+02
nc_qgeedgvrwglvme	vw_h61l96me3nfmwj	cl_yqr67t870n3viz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:12.766426+02	2022-05-18 11:24:12.766426+02
nc_zexmvjpsznxim1	vw_h61l96me3nfmwj	cl_vj4rthimqxnzck	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	18	2022-05-18 11:24:12.780364+02	2022-05-18 11:24:12.780364+02
nc_1wmwyju98irai3	vw_h61l96me3nfmwj	cl_ltyxan2j8mc59f	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	19	2022-05-18 11:24:12.789711+02	2022-05-18 11:24:12.789711+02
nc_9s70upt6og35ft	vw_h61l96me3nfmwj	cl_3s6cb84a61zbtv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	20	2022-05-18 11:24:12.795277+02	2022-05-18 11:24:12.795277+02
nc_fb1ryse5kz2u2y	vw_h61l96me3nfmwj	cl_s5uljkpg4meb33	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	21	2022-05-18 11:24:12.799606+02	2022-05-18 11:24:12.799606+02
nc_36v4u68mdfpat0	vw_h61l96me3nfmwj	cl_yr12t2sier2m7f	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	22	2022-05-18 11:24:12.803871+02	2022-05-18 11:24:12.803871+02
nc_pc56w2pnbgmbs0	vw_h61l96me3nfmwj	cl_flrhvi2qd887s8	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	23	2022-05-18 11:24:12.807016+02	2022-05-18 11:24:12.807016+02
nc_0geh0ja3l47joj	vw_h61l96me3nfmwj	cl_r2anw6zpujmv39	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	24	2022-05-18 11:24:12.811136+02	2022-05-18 11:24:12.811136+02
nc_369xd8qke9ahpk	vw_h61l96me3nfmwj	cl_n662hucflgmu9t	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	25	2022-05-18 11:24:12.815491+02	2022-05-18 11:24:12.815491+02
nc_w87y5gkk8922sn	vw_h61l96me3nfmwj	cl_mzo2rej6vzbu49	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	26	2022-05-18 11:24:12.819494+02	2022-05-18 11:24:12.819494+02
nc_fs00wbjhc7utp4	vw_h61l96me3nfmwj	cl_r6h0z44rb8rmgu	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	27	2022-05-18 11:24:12.823017+02	2022-05-18 11:24:12.823017+02
nc_1y7qu4lfja9idu	vw_h61l96me3nfmwj	cl_mkav6fudiwsrvy	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	28	2022-05-18 11:24:12.826639+02	2022-05-18 11:24:12.826639+02
nc_ocd1h430u2uny5	vw_h61l96me3nfmwj	cl_r7slbxotamjxlq	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	29	2022-05-18 11:24:12.831539+02	2022-05-18 11:24:12.831539+02
nc_egb007lkb3bgq7	vw_h61l96me3nfmwj	cl_hmy4b3d07hbwwm	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	30	2022-05-18 11:24:12.835045+02	2022-05-18 11:24:12.835045+02
nc_yatfxeyroflxwk	vw_h61l96me3nfmwj	cl_nweuxln7xxbqdz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	31	2022-05-18 11:24:12.838281+02	2022-05-18 11:24:12.838281+02
nc_9cmgv3l5m7600y	vw_h61l96me3nfmwj	cl_17imylhradkyjr	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	32	2022-05-18 11:24:12.841152+02	2022-05-18 11:24:12.841152+02
nc_zgn7ljtjpdo2ig	vw_h61l96me3nfmwj	cl_gn8fy43v6ost0s	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	33	2022-05-18 11:24:12.8461+02	2022-05-18 11:24:12.8461+02
nc_o4v2pk4c7ygof6	vw_tsu33qayws91ma	cl_go4hvqdi86lz0l	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.940347+02	2022-05-18 11:24:12.940347+02
nc_4xlkiib0frac1m	vw_jq0zcol66s3voo	cl_jmagj4k4jx5ffw	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	12	2022-05-18 11:24:12.942746+02	2022-05-18 11:24:12.942746+02
nc_hge3jh2su15vjf	vw_u2jzjq9smernol	cl_h44rjic2ztmxeb	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	19	2022-05-18 11:24:12.946511+02	2022-05-18 11:24:12.946511+02
nc_9acgri0k3beaun	vw_t8o4jwxtrx9nf0	cl_9q88yelvrwjb6u	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:12.948833+02	2022-05-18 11:24:12.948833+02
nc_m0hfeohkpjan2c	vw_xfg9ot5ntsq77q	cl_63p3m5hmmo1sal	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:13.028832+02	2022-05-18 11:24:13.028832+02
nc_pbecgr8duz9yir	vw_sxajywo4jx9ml5	cl_8bq9golhirsif4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	10	2022-05-18 11:24:13.030873+02	2022-05-18 11:24:13.030873+02
nc_lsfsh4ivducc6j	vw_bg9irqfa68gtl3	cl_5szjsumbk0yl6e	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:13.033156+02	2022-05-18 11:24:13.033156+02
nc_kzcr4c358vty7c	vw_srnlvrf0gn6rul	cl_aedmuthbh2dyrz	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:13.035545+02	2022-05-18 11:24:13.035545+02
nc_9auhv7248hf4xs	vw_ewjjjsi3mbc5h9	cl_5o0i7pgrh2i5oh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	21	2022-05-18 11:24:13.088085+02	2022-05-18 11:24:13.088085+02
nc_cogiayz1yer3gl	vw_t8o4jwxtrx9nf0	cl_s0w0vr1psqp3jm	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	20	2022-05-18 11:24:13.115667+02	2022-05-18 11:24:13.115667+02
nc_tvmzak4r9c06d5	vw_t8o4jwxtrx9nf0	cl_maydm3rcbofz04	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	21	2022-05-18 11:24:13.132835+02	2022-05-18 11:24:13.132835+02
nc_an7ycpcj8icxsn	vw_t8o4jwxtrx9nf0	cl_p1fuzr7pw5w24g	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	22	2022-05-18 11:24:13.142846+02	2022-05-18 11:24:13.142846+02
nc_os4cbd3ofdsgz4	vw_t8o4jwxtrx9nf0	cl_vo3yhpdpc37tle	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	23	2022-05-18 11:24:13.153853+02	2022-05-18 11:24:13.153853+02
nc_5qk6ed4v6ffds8	vw_t8o4jwxtrx9nf0	cl_nv3xs219vjcf81	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	24	2022-05-18 11:24:13.16399+02	2022-05-18 11:24:13.16399+02
nc_wagcxz0cnbvmox	vw_avtdalhk6frgn9	cl_g9g36fknoddzj2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:12.766222+02	2022-05-18 11:24:12.766222+02
nc_7f1op723gogout	vw_avtdalhk6frgn9	cl_5em2fkdhoklrs5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	18	2022-05-18 11:24:12.777464+02	2022-05-18 11:24:12.777464+02
nc_q441a9jak9rsoh	vw_avtdalhk6frgn9	cl_qwyxc7u35tff2f	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	19	2022-05-18 11:24:12.78745+02	2022-05-18 11:24:12.78745+02
nc_f00x6zfz4ysvxc	vw_om33x8ttx67est	cl_a7kx6brjog1as9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	5	2022-05-18 11:24:12.937003+02	2022-05-18 11:24:12.937003+02
nc_hgbe42c6l725oc	vw_sxajywo4jx9ml5	cl_art81q63geekql	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:12.938989+02	2022-05-18 11:24:12.938989+02
nc_vs7pxpts2nbivw	vw_ipcsobcohj1b6c	cl_6fou6htyzx6fa8	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	8	2022-05-18 11:24:12.941163+02	2022-05-18 11:24:12.941163+02
nc_hve72e65m2wam4	vw_bg9irqfa68gtl3	cl_sxpcp2cwk39cag	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:12.943922+02	2022-05-18 11:24:12.943922+02
nc_2eo2sksnedrcmp	vw_srnlvrf0gn6rul	cl_fzz2wpiidlwgie	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:12.947324+02	2022-05-18 11:24:12.947324+02
nc_3kf5j37e1b9iah	vw_t8o4jwxtrx9nf0	cl_lm7p4u967pdvrv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	18	2022-05-18 11:24:13.03885+02	2022-05-18 11:24:13.03885+02
nc_e7xsyml93ru9la	vw_sxajywo4jx9ml5	cl_iih5ukipx3n3mn	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:13.086732+02	2022-05-18 11:24:13.086732+02
nc_8aifl45nguuuhh	vw_bg9irqfa68gtl3	cl_xnqmpmsx7xxspc	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	19	2022-05-18 11:24:13.112979+02	2022-05-18 11:24:13.112979+02
nc_wuy8883ifygnuy	vw_ztmg51h4hma3hj	cl_oe5982qxpft694	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	12	2022-05-18 11:24:13.036522+02	2022-05-18 11:30:30.873795+02
nc_jtt7f25hogdoue	vw_ztmg51h4hma3hj	cl_2jgt0lwyo55epa	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	13	2022-05-18 11:24:13.088883+02	2022-05-18 11:30:32.191302+02
nc_8z01yisxs2oyus	vw_du7zrxhj51z27u	cl_dic1zbej0ei7x4	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	3	2022-05-18 11:24:12.938796+02	2022-05-18 11:24:12.938796+02
nc_nf2cipyditxzf3	vw_n362p00huzpsod	cl_j44zcituyfu32m	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.940936+02	2022-05-18 11:24:12.940936+02
nc_xqdil2genvb8ym	vw_6exoiii01n21je	cl_7s13f8gdnr3s0p	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	13	2022-05-18 11:24:12.943359+02	2022-05-18 11:24:12.943359+02
nc_s356vr2bu6j1jz	vw_352nd6p9zohj5w	cl_hzx263qp16kmx1	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:13.033851+02	2022-05-18 11:24:13.033851+02
nc_fmvf8thchrcbqv	vw_7ppnzs3y6eip9i	cl_l0q6rlkn5vn5ig	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	11	2022-05-18 11:24:13.036172+02	2022-05-18 11:24:13.036172+02
nc_oyi5b0xv2pb5bc	vw_hk5zatfifol9yl	cl_6nefo9nk04gtus	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	16	2022-05-18 11:24:13.038596+02	2022-05-18 11:24:13.038596+02
nc_uqeozsrdoafq6r	vw_srnlvrf0gn6rul	cl_f3nxqy328zxk7v	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	18	2022-05-18 11:24:13.088611+02	2022-05-18 11:24:13.088611+02
nc_gzewcqen6r4g5r	vw_hk5zatfifol9yl	cl_ly7t4ezyypj5ld	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	18	2022-05-18 11:24:13.112705+02	2022-05-18 11:24:13.112705+02
nc_6etmnjegst2qyq	vw_hk5zatfifol9yl	cl_e2mhvq3w5in28m	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	19	2022-05-18 11:24:13.132149+02	2022-05-18 11:24:13.132149+02
nc_5lk7x0clfw3i9f	vw_29tn3xki80n7e3	cl_wzx2z9fw9kc75l	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	15	2022-05-18 11:24:12.947106+02	2022-05-18 11:56:38.998077+02
nc_ef2usi7ryiwn9z	vw_ea8mjggb90pnnt	cl_9rx74ga8teuo7x	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	23	2022-05-18 11:24:12.951718+02	2022-05-18 11:24:12.951718+02
nc_uwvwtpsla33lbh	vw_mxy2on3nt3kj0g	cl_yuvf9ho2j6gyx2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:13.029115+02	2022-05-18 11:24:13.029115+02
nc_tr8hkmri9q7wus	vw_go1u33h41idruk	cl_xm8dld9std1g53	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:13.031129+02	2022-05-18 11:24:13.031129+02
nc_4viy5sudtg6pxh	vw_ipcsobcohj1b6c	cl_6dhtrez796pgcx	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	9	2022-05-18 11:24:13.033567+02	2022-05-18 11:24:13.033567+02
nc_vtcfqth8qej9id	vw_etd91poe0pvyld	cl_l7p1sgvsfxg2m2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	18	2022-05-18 11:24:13.035951+02	2022-05-18 11:24:13.035951+02
nc_zs795h2eyxkp1z	vw_51d0c7bmbclwb8	cl_8jwbbixyl9l1gu	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:13.038369+02	2022-05-18 11:24:13.038369+02
nc_1vf809n0qd8tiu	vw_jbveqkpzj241da	cl_38nvbo6hofo1ba	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	17	2022-05-18 11:24:13.08636+02	2022-05-18 11:24:13.08636+02
nc_f6hu9pyxzxjh4c	vw_u2jzjq9smernol	cl_9zapplyeoqzc3w	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	21	2022-05-18 11:24:13.08838+02	2022-05-18 11:24:13.08838+02
nc_3aex784a2mkqf3	vw_h61l96me3nfmwj	cl_j5ssutoh7jpsjp	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	37	2022-05-18 11:24:13.115933+02	2022-05-18 11:24:13.115933+02
nc_qv2adrudp4dzdb	vw_h61l96me3nfmwj	cl_wjjizc3l97gv2y	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	38	2022-05-18 11:24:13.133093+02	2022-05-18 11:24:13.133093+02
nc_9byc1on92rkjac	vw_h61l96me3nfmwj	cl_nhb4fulxib5qgn	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	39	2022-05-18 11:24:13.14719+02	2022-05-18 11:24:13.14719+02
nc_qsiv1lrr3fismv	vw_h61l96me3nfmwj	cl_ll6jzhxjga93yn	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	40	2022-05-18 11:24:13.164492+02	2022-05-18 11:24:13.164492+02
nc_gri6zu839pyp04	vw_h61l96me3nfmwj	cl_nht4pi4vgcjvr2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	41	2022-05-18 11:24:13.171983+02	2022-05-18 11:24:13.171983+02
nc_94f9toska475n6	vw_h61l96me3nfmwj	cl_s55zpov56wgfmh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	42	2022-05-18 11:24:13.17967+02	2022-05-18 11:24:13.17967+02
nc_pcinjdvpanobya	vw_h61l96me3nfmwj	cl_200i4b3amjhxyg	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	43	2022-05-18 11:24:13.186172+02	2022-05-18 11:24:13.186172+02
nc_6gua3ex81qv5m0	vw_h61l96me3nfmwj	cl_stxhadl7xlmpqv	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	44	2022-05-18 11:24:13.191342+02	2022-05-18 11:24:13.191342+02
nc_zp5huyhji8enro	vw_h61l96me3nfmwj	cl_meczg88bpm9dk2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	45	2022-05-18 11:24:13.198653+02	2022-05-18 11:24:13.198653+02
nc_c8zs2qaf191uh9	vw_h61l96me3nfmwj	cl_s0x15kjbe4c2at	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	46	2022-05-18 11:24:13.204187+02	2022-05-18 11:24:13.204187+02
nc_fagely39mueyc9	vw_h61l96me3nfmwj	cl_et3x24qyhtam8z	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	47	2022-05-18 11:24:13.210228+02	2022-05-18 11:24:13.210228+02
nc_7qagfpst4l0g9i	vw_h61l96me3nfmwj	cl_iaf4ewburqa3fd	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	48	2022-05-18 11:24:13.217341+02	2022-05-18 11:24:13.217341+02
nc_brg72gdn5lltcr	vw_h61l96me3nfmwj	cl_ob4dqzwkr599rk	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	49	2022-05-18 11:24:13.223317+02	2022-05-18 11:24:13.223317+02
nc_rwjznej7o38fda	vw_h61l96me3nfmwj	cl_83lo23d61k9mdi	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	50	2022-05-18 11:24:13.230821+02	2022-05-18 11:24:13.230821+02
nc_wx1r1yq5kudt22	vw_h61l96me3nfmwj	cl_h1zilufq67bh99	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	51	2022-05-18 11:24:13.236974+02	2022-05-18 11:24:13.236974+02
nc_2pts3u623s4r6y	vw_h61l96me3nfmwj	cl_lbsn3zcyhy5knh	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	52	2022-05-18 11:24:13.243311+02	2022-05-18 11:24:13.243311+02
nc_ywiemgwqgfqag8	vw_h61l96me3nfmwj	cl_d13oki8bjyop85	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	53	2022-05-18 11:24:13.250712+02	2022-05-18 11:24:13.250712+02
nc_wknf8e9gqoylup	vw_h61l96me3nfmwj	cl_kh8758rfbidubl	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	54	2022-05-18 11:24:13.256263+02	2022-05-18 11:24:13.256263+02
nc_q2jco19gs5eeom	vw_c525afyrdh8nqn	cl_xzi1g0bfxea38l	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	4	2022-05-18 11:24:13.263794+02	2022-05-18 11:24:13.263794+02
nc_1njsk7y09p3324	vw_tsu33qayws91ma	cl_xvgp8vm7xginj2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	14	2022-05-18 11:24:13.26882+02	2022-05-18 11:24:13.26882+02
nc_upjehdcg4bbeqj	vw_u2jzjq9smernol	cl_n0gu1q86ugzp2h	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	23	2022-05-18 11:24:13.279784+02	2022-05-18 11:24:13.279784+02
nc_556rfc6o3ik2qo	vw_81imke2v3e15l5	cl_hbp7w3tjwejjua	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	6	2022-05-18 11:24:13.284992+02	2022-05-18 11:24:13.284992+02
nc_gjss8z6lhxrtn4	vw_q1cl9wv3hhuvca	cl_b7z0ydp2g1nxu9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	7	2022-05-18 11:24:13.301154+02	2022-05-18 11:24:13.301154+02
nc_y7e586bhzk52qg	vw_dejssdomuxg3uw	cl_ezsdkn1opttaz3	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	20	2022-05-18 11:24:13.3061+02	2022-05-18 11:24:13.3061+02
nc_aix0jqtvuekc9d	vw_andyomdmvq8tv6	cl_i8optm3oaj6xqp	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	19	2022-05-18 11:24:13.089095+02	2022-05-18 11:24:13.089095+02
nc_6a5vlsaoh4sry6	vw_ewjjjsi3mbc5h9	cl_ub0vo7g9n0zjrb	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	22	2022-05-18 11:24:13.113209+02	2022-05-18 11:24:13.113209+02
nc_0z1k4ez4tylydt	vw_ewjjjsi3mbc5h9	cl_7s2mqpkdq9y4f5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	23	2022-05-18 11:24:13.132464+02	2022-05-18 11:24:13.132464+02
nc_6hyt8a3oo2hgtv	vw_ewjjjsi3mbc5h9	cl_9xl9rhlhzm8ga2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	24	2022-05-18 11:24:13.146914+02	2022-05-18 11:24:13.146914+02
nc_8kecusz3j5imwf	vw_ewjjjsi3mbc5h9	cl_5u2tjgd4jktqj2	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	25	2022-05-18 11:24:13.159401+02	2022-05-18 11:24:13.159401+02
nc_kr4phk5md0vvqr	vw_ztmg51h4hma3hj	cl_68x1go22132wqm	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	137px	t	1	2022-05-18 11:24:11.932387+02	2022-05-18 11:29:27.373022+02
nc_8akjzb1uhrsdcj	vw_29tn3xki80n7e3	cl_n2appwjzrymdk5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	10	2022-05-18 11:24:12.589623+02	2022-05-18 11:56:38.998077+02
nc_h4jv3cjxwol3b7	vw_29tn3xki80n7e3	cl_ej7h0ko9o7l40k	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	11	2022-05-18 11:24:12.631771+02	2022-05-18 11:56:38.998077+02
nc_j4c7hu25zdc28r	vw_29tn3xki80n7e3	cl_min1a7d8k76n5i	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	6	2022-05-18 11:24:12.349894+02	2022-05-18 11:56:38.998077+02
nc_d1b0rk1elplhze	vw_29tn3xki80n7e3	cl_lxtgajo4s3br49	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	7	2022-05-18 11:24:12.421394+02	2022-05-18 11:56:38.998077+02
nc_1ctl4eiir4hman	vw_29tn3xki80n7e3	cl_wo85353duh2q82	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	8	2022-05-18 11:24:12.484025+02	2022-05-18 11:56:38.998077+02
nc_ddktpy6477yxxb	vw_29tn3xki80n7e3	cl_k2zpyuzl42o0ec	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	12	2022-05-18 11:24:12.671623+02	2022-05-18 11:56:38.998077+02
nc_dveo2yxpazfo7a	vw_29tn3xki80n7e3	cl_03jw5hlwoiw2ji	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	f	17	2022-05-18 11:24:13.086907+02	2022-05-18 11:56:38.998077+02
nc_be8vw1d0hnai0t	vw_29tn3xki80n7e3	cl_pjm18dcg199nid	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	\N	\N	200px	t	19	2022-05-18 11:58:47.534257+02	2022-05-18 11:58:47.534257+02
\.


--
-- Data for Name: nc_grid_view_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_grid_view_v2 (fk_view_id, base_id, project_id, uuid, created_at, updated_at) FROM stdin;
vw_om33x8ttx67est	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.686613+02	2022-05-18 11:24:11.686613+02
vw_xfg9ot5ntsq77q	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.688364+02	2022-05-18 11:24:11.688364+02
vw_c525afyrdh8nqn	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.688626+02	2022-05-18 11:24:11.688626+02
vw_tsu33qayws91ma	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.704289+02	2022-05-18 11:24:11.704289+02
vw_mxy2on3nt3kj0g	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.709808+02	2022-05-18 11:24:11.709808+02
vw_jq0zcol66s3voo	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.721622+02	2022-05-18 11:24:11.721622+02
vw_ipcsobcohj1b6c	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.722199+02	2022-05-18 11:24:11.722199+02
vw_h61l96me3nfmwj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.723828+02	2022-05-18 11:24:11.723828+02
vw_u2jzjq9smernol	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.724075+02	2022-05-18 11:24:11.724075+02
vw_2lgkfm9g3xygp1	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.724424+02	2022-05-18 11:24:11.724424+02
vw_ewjjjsi3mbc5h9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.724824+02	2022-05-18 11:24:11.724824+02
vw_dzgaxceud97gqt	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.725072+02	2022-05-18 11:24:11.725072+02
vw_jbveqkpzj241da	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.725447+02	2022-05-18 11:24:11.725447+02
vw_gl8cwxr6m3iq38	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.725808+02	2022-05-18 11:24:11.725808+02
vw_sxajywo4jx9ml5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.72639+02	2022-05-18 11:24:11.72639+02
vw_hk5zatfifol9yl	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.727261+02	2022-05-18 11:24:11.727261+02
vw_pcdpz0lqr00s9k	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.729713+02	2022-05-18 11:24:11.729713+02
vw_etd91poe0pvyld	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.735414+02	2022-05-18 11:24:11.735414+02
vw_n362p00huzpsod	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.736441+02	2022-05-18 11:24:11.736441+02
vw_wflzr4fus2ngby	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.738769+02	2022-05-18 11:24:11.738769+02
vw_andyomdmvq8tv6	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.740094+02	2022-05-18 11:24:11.740094+02
vw_7fueohglirse69	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.740422+02	2022-05-18 11:24:11.740422+02
vw_go1u33h41idruk	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.740702+02	2022-05-18 11:24:11.740702+02
vw_avtdalhk6frgn9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.741061+02	2022-05-18 11:24:11.741061+02
vw_ea8mjggb90pnnt	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.741446+02	2022-05-18 11:24:11.741446+02
vw_6exoiii01n21je	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.743677+02	2022-05-18 11:24:11.743677+02
vw_352nd6p9zohj5w	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.748938+02	2022-05-18 11:24:11.748938+02
vw_lalt4t2bjas0gj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.75145+02	2022-05-18 11:24:11.75145+02
vw_bg9irqfa68gtl3	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.753308+02	2022-05-18 11:24:11.753308+02
vw_q1cl9wv3hhuvca	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.753771+02	2022-05-18 11:24:11.753771+02
vw_brn7rul3d130ex	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.754102+02	2022-05-18 11:24:11.754102+02
vw_7ppnzs3y6eip9i	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.755416+02	2022-05-18 11:24:11.755416+02
vw_51d0c7bmbclwb8	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.757502+02	2022-05-18 11:24:11.757502+02
vw_lr57pe50bbmjar	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.75841+02	2022-05-18 11:24:11.75841+02
vw_mauaaqk5b7v0rr	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.759249+02	2022-05-18 11:24:11.759249+02
vw_t8o4jwxtrx9nf0	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.759639+02	2022-05-18 11:24:11.759639+02
vw_dejssdomuxg3uw	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.767908+02	2022-05-18 11:24:11.767908+02
vw_29tn3xki80n7e3	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.77294+02	2022-05-18 11:24:11.77294+02
vw_ztmg51h4hma3hj	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.775705+02	2022-05-18 11:24:11.775705+02
vw_du7zrxhj51z27u	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.776354+02	2022-05-18 11:24:11.776354+02
vw_ccggsien3mx2nl	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.777108+02	2022-05-18 11:24:11.777108+02
vw_5w6mg5ux0ryjjl	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.779382+02	2022-05-18 11:24:11.779382+02
vw_srnlvrf0gn6rul	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.780459+02	2022-05-18 11:24:11.780459+02
vw_r15acq81tqlkxp	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.783532+02	2022-05-18 11:24:11.783532+02
vw_81imke2v3e15l5	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.783934+02	2022-05-18 11:24:11.783934+02
vw_idcyonon9t2t5l	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.784852+02	2022-05-18 11:24:11.784852+02
vw_yrmq5jknxj9sx6	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.785236+02	2022-05-18 11:24:11.785236+02
vw_8kwir8e38t3fc9	ds_aej9ao4mzw9cff	p_41297240e6of48	\N	2022-05-18 11:24:11.794834+02	2022-05-18 11:24:11.794834+02
\.


--
-- Data for Name: nc_hook_logs_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_hook_logs_v2 (id, base_id, project_id, fk_hook_id, type, event, operation, test_call, payload, conditions, notification, error_code, error_message, error, execution_time, response, triggered_by, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_hooks; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_hooks (id, project_id, db_alias, title, description, env, tn, type, event, operation, async, payload, url, headers, condition, notification, retries, retry_interval, timeout, active, created_at, updated_at) FROM stdin;
1	\N	db	\N	\N	all	\N	AUTH_MIDDLEWARE	\N	\N	f	t	\N	\N	\N	\N	0	60000	60000	t	\N	\N
\.


--
-- Data for Name: nc_hooks_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_hooks_v2 (id, base_id, project_id, fk_model_id, title, description, env, type, event, operation, async, payload, url, headers, condition, notification, retries, retry_interval, timeout, active, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_kanban_view_columns_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_kanban_view_columns_v2 (id, base_id, project_id, fk_view_id, fk_column_id, uuid, label, help, show, "order", created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_kanban_view_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_kanban_view_v2 (fk_view_id, base_id, project_id, show, "order", uuid, title, public, password, show_all_fields, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_loaders; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_loaders (id, project_id, db_alias, title, parent, child, relation, resolver, functions, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_migrations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_migrations (id, project_id, db_alias, up, down, title, title_down, description, batch, checksum, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_models; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_models (id, project_id, db_alias, title, alias, type, meta, schema, schema_previous, services, messages, enabled, parent_model_title, show_as, query_params, list_idx, tags, pinned, created_at, updated_at, mm, m_to_m_meta, "order", view_order) FROM stdin;
\.


--
-- Data for Name: nc_models_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_models_v2 (id, base_id, project_id, table_name, title, type, meta, schema, enabled, mm, tags, pinned, deleted, "order", created_at, updated_at) FROM stdin;
md_79w8haiknxbkzd	ds_aej9ao4mzw9cff	p_41297240e6of48	categories	Categories	table	\N	\N	t	f	\N	\N	\N	2	2022-05-18 11:24:11.136919+02	2022-05-18 11:24:11.136919+02
md_71j5r6adaq9l3i	ds_aej9ao4mzw9cff	p_41297240e6of48	customer_demographics	CustomerDemographics	table	\N	\N	t	f	\N	\N	\N	4	2022-05-18 11:24:11.25584+02	2022-05-18 11:24:11.25584+02
md_vb4nqekkid2rl6	ds_aej9ao4mzw9cff	p_41297240e6of48	customers	Customers	table	\N	\N	t	f	\N	\N	\N	5	2022-05-18 11:24:11.258134+02	2022-05-18 11:24:11.258134+02
md_yi49xujy865heu	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_bases_v2	NcBasesV2	table	\N	\N	t	f	\N	\N	\N	9	2022-05-18 11:24:11.434418+02	2022-05-18 11:24:11.434418+02
md_5hjnvbjhgfiz6r	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_col_lookup_v2	NcColLookupV2	table	\N	\N	t	f	\N	\N	\N	11	2022-05-18 11:24:11.434956+02	2022-05-18 11:24:11.434956+02
md_tm1aur9f3yn2l2	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_col_formula_v2	NcColFormulaV2	table	\N	\N	t	f	\N	\N	\N	10	2022-05-18 11:24:11.43542+02	2022-05-18 11:24:11.43542+02
md_ue8cshscw72tz4	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_columns_v2	NcColumnsV2	table	\N	\N	t	f	\N	\N	\N	15	2022-05-18 11:24:11.451859+02	2022-05-18 11:24:11.451859+02
md_yor91fufon7h6c	ds_aej9ao4mzw9cff	p_41297240e6of48	employees	Employees	table	\N	\N	t	f	\N	\N	\N	7	2022-05-18 11:24:11.453207+02	2022-05-18 11:24:11.453207+02
md_91fcdvmtveuz77	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_disabled_models_for_role_v2	NcDisabledModelsForRoleV2	table	\N	\N	t	f	\N	\N	\N	16	2022-05-18 11:24:11.454173+02	2022-05-18 11:24:11.454173+02
md_ruvc4yz1y8yonz	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_col_relations_v2	NcColRelationsV2	table	\N	\N	t	f	\N	\N	\N	12	2022-05-18 11:24:11.455494+02	2022-05-18 11:24:11.455494+02
md_rmg6p4gh772ip8	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_col_rollup_v2	NcColRollupV2	table	\N	\N	t	f	\N	\N	\N	13	2022-05-18 11:24:11.458395+02	2022-05-18 11:24:11.458395+02
md_6pjpqivc2ymh3b	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_col_select_options_v2	NcColSelectOptionsV2	table	\N	\N	t	f	\N	\N	\N	14	2022-05-18 11:24:11.459315+02	2022-05-18 11:24:11.459315+02
md_hle53jhinff293	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_audit_v2	NcAuditV2	table	\N	\N	t	f	\N	\N	\N	8	2022-05-18 11:24:11.460673+02	2022-05-18 11:24:11.460673+02
md_gwspodfugzyzs2	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_filter_exp_v2	NcFilterExpV2	table	\N	\N	t	f	\N	\N	\N	17	2022-05-18 11:24:11.46739+02	2022-05-18 11:24:11.46739+02
md_ajkugjxa4haf22	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_kanban_view_v2	NcKanbanViewV2	table	\N	\N	t	f	\N	\N	\N	27	2022-05-18 11:24:11.471543+02	2022-05-18 11:24:11.471543+02
md_op8dq4us6zg5tn	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_form_view_columns_v2	NcFormViewColumnsV2	table	\N	\N	t	f	\N	\N	\N	18	2022-05-18 11:24:11.584783+02	2022-05-18 11:24:11.584783+02
md_47c5g91l1nc5se	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_form_view_v2	NcFormViewV2	table	\N	\N	t	f	\N	\N	\N	19	2022-05-18 11:24:11.586059+02	2022-05-18 11:24:11.586059+02
md_thf0xauk4e6qxs	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_grid_view_columns_v2	NcGridViewColumnsV2	table	\N	\N	t	f	\N	\N	\N	22	2022-05-18 11:24:11.58652+02	2022-05-18 11:24:11.58652+02
md_2vqez05sn215dj	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_gallery_view_v2	NcGalleryViewV2	table	\N	\N	t	f	\N	\N	\N	21	2022-05-18 11:24:11.586957+02	2022-05-18 11:24:11.586957+02
md_ejwlk09y393nhf	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_kanban_view_columns_v2	NcKanbanViewColumnsV2	table	\N	\N	t	f	\N	\N	\N	26	2022-05-18 11:24:11.589413+02	2022-05-18 11:24:11.589413+02
md_zmhiordb5vpwkf	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_grid_view_v2	NcGridViewV2	table	\N	\N	t	f	\N	\N	\N	23	2022-05-18 11:24:11.589909+02	2022-05-18 11:24:11.589909+02
md_whwvfz6cb3ljdb	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_gallery_view_columns_v2	NcGalleryViewColumnsV2	table	\N	\N	t	f	\N	\N	\N	20	2022-05-18 11:24:11.590207+02	2022-05-18 11:24:11.590207+02
md_mqny0xchyq3gqp	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_hook_logs_v2	NcHookLogsV2	table	\N	\N	t	f	\N	\N	\N	24	2022-05-18 11:24:11.59047+02	2022-05-18 11:24:11.59047+02
md_muq975426eve8k	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_hooks_v2	NcHooksV2	table	\N	\N	t	f	\N	\N	\N	25	2022-05-18 11:24:11.59074+02	2022-05-18 11:24:11.59074+02
md_ebe9bibmq1l2uf	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_project_users_v2	NcProjectUsersV2	table	\N	\N	t	f	\N	\N	\N	31	2022-05-18 11:24:11.638567+02	2022-05-18 11:24:11.638567+02
md_5ms9hq3t1a9cro	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_plugins_v2	NcPluginsV2	table	\N	\N	t	f	\N	\N	\N	30	2022-05-18 11:24:11.643736+02	2022-05-18 11:24:11.643736+02
md_ofa8d1ocm4oowx	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_projects_v2	NcProjectsV2	table	\N	\N	t	f	\N	\N	\N	32	2022-05-18 11:24:11.645391+02	2022-05-18 11:24:11.645391+02
md_nj2q6tfni3b5a8	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_orgs_v2	NcOrgsV2	table	\N	\N	t	f	\N	\N	\N	29	2022-05-18 11:24:11.646032+02	2022-05-18 11:24:11.646032+02
md_1wgrn82853mqsm	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_shared_bases	NcSharedBases	table	\N	\N	t	f	\N	\N	\N	33	2022-05-18 11:24:11.647243+02	2022-05-18 11:24:11.647243+02
md_0z5ofu8o9xph1j	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_teams_v2	NcTeamsV2	table	\N	\N	t	f	\N	\N	\N	37	2022-05-18 11:24:11.647987+02	2022-05-18 11:24:11.647987+02
md_wqy1a8c5ggpyea	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_sort_v2	NcSortV2	table	\N	\N	t	f	\N	\N	\N	35	2022-05-18 11:24:11.649579+02	2022-05-18 11:24:11.649579+02
md_12azccujctejop	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_shared_views_v2	NcSharedViewsV2	table	\N	\N	t	f	\N	\N	\N	34	2022-05-18 11:24:11.650795+02	2022-05-18 11:24:11.650795+02
md_cltahxtbesn8vs	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_models_v2	NcModelsV2	table	\N	\N	t	f	\N	\N	\N	28	2022-05-18 11:24:11.651816+02	2022-05-18 11:24:11.651816+02
md_6ht8y7vgzcek0p	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_users_v2	NcUsersV2	table	\N	\N	t	f	\N	\N	\N	38	2022-05-18 11:24:11.658179+02	2022-05-18 11:24:11.658179+02
md_mmp2o2atqtudsn	ds_aej9ao4mzw9cff	p_41297240e6of48	products	Products	table	\N	\N	t	f	\N	\N	\N	42	2022-05-18 11:24:11.668444+02	2022-05-18 11:24:11.668444+02
md_jgq6gbzvyfbk2g	ds_aej9ao4mzw9cff	p_41297240e6of48	region	Region	table	\N	\N	t	f	\N	\N	\N	43	2022-05-18 11:24:11.668903+02	2022-05-18 11:24:11.668903+02
md_s9jny9njrhdnwz	ds_aej9ao4mzw9cff	p_41297240e6of48	order_details	OrderDetails	table	\N	\N	t	f	\N	\N	\N	40	2022-05-18 11:24:11.669269+02	2022-05-18 11:24:11.669269+02
md_r7cgbvzh9qy7qk	ds_aej9ao4mzw9cff	p_41297240e6of48	orders	Orders	table	\N	\N	t	f	\N	\N	\N	41	2022-05-18 11:24:11.669768+02	2022-05-18 11:24:11.669768+02
md_eqrk8yd76gurbf	ds_aej9ao4mzw9cff	p_41297240e6of48	suppliers	Suppliers	table	\N	\N	t	f	\N	\N	\N	45	2022-05-18 11:24:11.674913+02	2022-05-18 11:24:11.674913+02
md_wk0cg5v1lnk9ew	ds_aej9ao4mzw9cff	p_41297240e6of48	us_states	UsStates	table	\N	\N	t	f	\N	\N	\N	47	2022-05-18 11:24:11.675347+02	2022-05-18 11:24:11.675347+02
md_dg3rmvfk4bsg1i	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_views_v2	NcViewsV2	table	\N	\N	t	f	\N	\N	\N	39	2022-05-18 11:24:11.675654+02	2022-05-18 11:24:11.675654+02
md_j409llgiqo90yz	ds_aej9ao4mzw9cff	p_41297240e6of48	territories	Territories	table	\N	\N	t	f	\N	\N	\N	46	2022-05-18 11:24:11.676006+02	2022-05-18 11:24:11.676006+02
md_wjrbh5du4ifq5b	ds_aej9ao4mzw9cff	p_41297240e6of48	shippers	Shippers	table	\N	\N	t	f	\N	\N	\N	44	2022-05-18 11:24:11.676587+02	2022-05-18 11:24:11.676587+02
md_llr0f8esceyorr	ds_aej9ao4mzw9cff	p_41297240e6of48	xc_knex_migrationsv2	XcKnexMigrationsv2	table	\N	\N	t	f	\N	\N	\N	48	2022-05-18 11:24:11.682254+02	2022-05-18 11:24:11.682254+02
md_c1y2qwn3okt75w	ds_aej9ao4mzw9cff	p_41297240e6of48	xc_knex_migrationsv2_lock	XcKnexMigrationsv2Lock	table	\N	\N	t	f	\N	\N	\N	49	2022-05-18 11:24:11.690597+02	2022-05-18 11:24:11.690597+02
md_7vpsxklk3sph75	ds_aej9ao4mzw9cff	p_41297240e6of48	customer_customer_demo	CustomerCustomerDemo	table	\N	\N	t	t	\N	\N	\N	3	2022-05-18 11:24:11.251835+02	2022-05-18 11:24:13.27033+02
md_q2z56jvps22ukx	ds_aej9ao4mzw9cff	p_41297240e6of48	employee_territories	EmployeeTerritories	table	\N	\N	t	t	\N	\N	\N	6	2022-05-18 11:24:11.332817+02	2022-05-18 11:24:13.286424+02
md_89v98o1rpi4idv	ds_aej9ao4mzw9cff	p_41297240e6of48	nc_team_users_v2	NcTeamUsersV2	table	\N	\N	t	t	\N	\N	\N	36	2022-05-18 11:24:11.643157+02	2022-05-18 11:24:13.307474+02
\.


--
-- Data for Name: nc_orgs_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_orgs_v2 (id, title, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_plugins; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_plugins (id, project_id, db_alias, title, description, active, rating, version, docs, status, status_details, logo, icon, tags, category, input_schema, input, creator, creator_website, price, created_at, updated_at) FROM stdin;
1	\N	\N	Google	Google OAuth2 login.	f	\N	0.0.1	\N	install	\N	plugins/google.png	\N	Authentication	Google	{"title":"Configure Google Auth","items":[{"key":"client_id","label":"Client ID","placeholder":"Client ID","type":"SingleLineText","required":true},{"key":"client_secret","label":"Client Secret","placeholder":"Client Secret","type":"Password","required":true},{"key":"redirect_url","label":"Redirect URL","placeholder":"Redirect URL","type":"SingleLineText","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and configured Google Authentication, restart NocoDB","msgOnUninstall":""}	\N	\N	\N	Free	\N	\N
3	\N	\N	Metadata LRU Cache	A cache object that deletes the least-recently-used items.	t	\N	0.0.1	\N	install	\N	plugins/xgene.png	\N	Cache	Cache	{"title":"Configure Metadata LRU Cache","items":[{"key":"max","label":"Maximum Size","placeholder":"Maximum Size","type":"SingleLineText","required":true},{"key":"maxAge","label":"Maximum Age(in ms)","placeholder":"Maximum Age(in ms)","type":"SingleLineText","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully updated LRU cache options.","msgOnUninstall":""}	{"max":500,"maxAge":86400000}	\N	\N	Free	\N	\N
\.


--
-- Data for Name: nc_plugins_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_plugins_v2 (id, title, description, active, rating, version, docs, status, status_details, logo, icon, tags, category, input_schema, input, creator, creator_website, price, created_at, updated_at) FROM stdin;
nc_xmkfu4vxjkx6km	Slack	Slack brings team communication and collaboration into one place so you can get more work done, whether you belong to a large enterprise or a small business. 	f	\N	0.0.1	\N	install	\N	plugins/slack.webp	\N	Chat	Chat	{"title":"Configure Slack","array":true,"items":[{"key":"channel","label":"Channel Name","placeholder":"Channel Name","type":"SingleLineText","required":true},{"key":"webhook_url","label":"Webhook URL","placeholder":"Webhook URL","type":"Password","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and Slack is enabled for notification.","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.65496+02	2022-05-18 11:22:25.65496+02
nc_02vuykp17tg45r	Microsoft Teams	Microsoft Teams is for everyone · Instantly go from group chat to video call with the touch of a button.	f	\N	0.0.1	\N	install	\N	plugins/teams.ico	\N	Chat	Chat	{"title":"Configure Microsoft Teams","array":true,"items":[{"key":"channel","label":"Channel Name","placeholder":"Channel Name","type":"SingleLineText","required":true},{"key":"webhook_url","label":"Webhook URL","placeholder":"Webhook URL","type":"Password","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and Microsoft Teams is enabled for notification.","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.657166+02	2022-05-18 11:22:25.657166+02
nc_z8mg6uhc7l47ye	Discord	Discord is the easiest way to talk over voice, video, and text. Talk, chat, hang out, and stay close with your friends and communities.	f	\N	0.0.1	\N	install	\N	plugins/discord.png	\N	Chat	Chat	{"title":"Configure Discord","array":true,"items":[{"key":"channel","label":"Channel Name","placeholder":"Channel Name","type":"SingleLineText","required":true},{"key":"webhook_url","label":"Webhook URL","type":"Password","placeholder":"Webhook URL","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and Discord is enabled for notification.","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.659408+02	2022-05-18 11:22:25.659408+02
nc_3silfnb7x2xfve	Whatsapp Twilio	With Twilio, unite communications and strengthen customer relationships across your business – from marketing and sales to customer service and operations.	f	\N	0.0.1	\N	install	\N	plugins/whatsapp.png	\N	Chat	Twilio	{"title":"Configure Twilio","items":[{"key":"sid","label":"Account SID","placeholder":"Account SID","type":"SingleLineText","required":true},{"key":"token","label":"Auth Token","placeholder":"Auth Token","type":"Password","required":true},{"key":"from","label":"From Phone Number","placeholder":"From Phone Number","type":"SingleLineText","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and Whatsapp Twilio is enabled for notification.","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.661023+02	2022-05-18 11:22:25.661023+02
nc_qtl1d27mulx0w2	Twilio	With Twilio, unite communications and strengthen customer relationships across your business – from marketing and sales to customer service and operations.	f	\N	0.0.1	\N	install	\N	plugins/twilio.png	\N	Chat	Twilio	{"title":"Configure Twilio","items":[{"key":"sid","label":"Account SID","placeholder":"Account SID","type":"SingleLineText","required":true},{"key":"token","label":"Auth Token","placeholder":"Auth Token","type":"Password","required":true},{"key":"from","label":"From Phone Number","placeholder":"From Phone Number","type":"SingleLineText","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and Twilio is enabled for notification.","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.662461+02	2022-05-18 11:22:25.662461+02
nc_1dshjhh9m4hob3	S3	Amazon Simple Storage Service (Amazon S3) is an object storage service that offers industry-leading scalability, data availability, security, and performance.	f	\N	0.0.1	\N	install	\N	plugins/s3.png	\N	Storage	Storage	{"title":"Configure Amazon S3","items":[{"key":"bucket","label":"Bucket Name","placeholder":"Bucket Name","type":"SingleLineText","required":true},{"key":"region","label":"Region","placeholder":"Region","type":"SingleLineText","required":true},{"key":"access_key","label":"Access Key","placeholder":"Access Key","type":"SingleLineText","required":true},{"key":"access_secret","label":"Access Secret","placeholder":"Access Secret","type":"Password","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and attachment will be stored in AWS S3","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.663824+02	2022-05-18 11:22:25.663824+02
nc_pgcz9d5ej531ci	Minio	MinIO is a High Performance Object Storage released under Apache License v2.0. It is API compatible with Amazon S3 cloud storage service.	f	\N	0.0.1	\N	install	\N	plugins/minio.png	\N	Storage	Storage	{"title":"Configure Minio","items":[{"key":"endPoint","label":"Minio Endpoint","placeholder":"Minio Endpoint","type":"SingleLineText","required":true},{"key":"port","label":"Port","placeholder":"Port","type":"Number","required":true},{"key":"bucket","label":"Bucket Name","placeholder":"Bucket Name","type":"SingleLineText","required":true},{"key":"access_key","label":"Access Key","placeholder":"Access Key","type":"SingleLineText","required":true},{"key":"access_secret","label":"Access Secret","placeholder":"Access Secret","type":"Password","required":true},{"key":"useSSL","label":"Use SSL","placeholder":"Use SSL","type":"Checkbox","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and attachment will be stored in Minio","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.665067+02	2022-05-18 11:22:25.665067+02
nc_gs7fe2rycs5i4o	GCS	Google Cloud Storage is a RESTful online file storage web service for storing and accessing data on Google Cloud Platform infrastructure.	f	\N	0.0.1	\N	install	\N	plugins/gcs.png	\N	Storage	Storage	{"title":"Configure Google Cloud Storage","items":[{"key":"bucket","label":"Bucket Name","placeholder":"Bucket Name","type":"SingleLineText","required":true},{"key":"client_email","label":"Client Email","placeholder":"Client Email","type":"SingleLineText","required":true},{"key":"private_key","label":"Private Key","placeholder":"Private Key","type":"Password","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and attachment will be stored in Google Cloud Storage","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.666478+02	2022-05-18 11:22:25.666478+02
nc_8jz6bw6e8dtdgy	Mattermost	Mattermost brings all your team communication into one place, making it searchable and accessible anywhere.	f	\N	0.0.1	\N	install	\N	plugins/mattermost.png	\N	Chat	Chat	{"title":"Configure Mattermost","array":true,"items":[{"key":"channel","label":"Channel Name","placeholder":"Channel Name","type":"SingleLineText","required":true},{"key":"webhook_url","label":"Webhook URL","placeholder":"Webhook URL","type":"Password","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and Mattermost is enabled for notification.","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.667915+02	2022-05-18 11:22:25.667915+02
nc_ccrbc5j9wf4xge	Spaces	Store & deliver vast amounts of content with a simple architecture.	f	\N	0.0.1	\N	install	\N	plugins/spaces.png	\N	Storage	Storage	{"title":"DigitalOcean Spaces","items":[{"key":"bucket","label":"Bucket Name","placeholder":"Bucket Name","type":"SingleLineText","required":true},{"key":"region","label":"Region","placeholder":"Region","type":"SingleLineText","required":true},{"key":"access_key","label":"Access Key","placeholder":"Access Key","type":"SingleLineText","required":true},{"key":"access_secret","label":"Access Secret","placeholder":"Access Secret","type":"Password","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and attachment will be stored in DigitalOcean Spaces","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.669773+02	2022-05-18 11:22:25.669773+02
nc_jyk8kicnnte6c7	Backblaze B2	Backblaze B2 is enterprise-grade, S3 compatible storage that companies around the world use to store and serve data while improving their cloud OpEx vs. Amazon S3 and others.	f	\N	0.0.1	\N	install	\N	plugins/backblaze.jpeg	\N	Storage	Storage	{"title":"Configure Backblaze B2","items":[{"key":"bucket","label":"Bucket Name","placeholder":"Bucket Name","type":"SingleLineText","required":true},{"key":"region","label":"Region","placeholder":"Region","type":"SingleLineText","required":true},{"key":"access_key","label":"Access Key","placeholder":"Access Key","type":"SingleLineText","required":true},{"key":"access_secret","label":"Access Secret","placeholder":"Access Secret","type":"Password","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and attachment will be stored in Backblaze B2","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.671278+02	2022-05-18 11:22:25.671278+02
nc_kuamt7ym4ds2nc	Vultr Object Storage	Using Vultr Object Storage can give flexibility and cloud storage that allows applications greater flexibility and access worldwide.	f	\N	0.0.1	\N	install	\N	plugins/vultr.png	\N	Storage	Storage	{"title":"Configure Vultr Object Storage","items":[{"key":"bucket","label":"Bucket Name","placeholder":"Bucket Name","type":"SingleLineText","required":true},{"key":"access_key","label":"Access Key","placeholder":"Access Key","type":"SingleLineText","required":true},{"key":"access_secret","label":"Access Secret","placeholder":"Access Secret","type":"Password","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and attachment will be stored in Vultr Object Storage","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.672836+02	2022-05-18 11:22:25.672836+02
nc_oz6u5a4u9jnkbp	OvhCloud Object Storage	Upload your files to a space that you can access via HTTPS using the OpenStack Swift API, or the S3 API. 	f	\N	0.0.1	\N	install	\N	plugins/ovhCloud.png	\N	Storage	Storage	{"title":"Configure OvhCloud Object Storage","items":[{"key":"bucket","label":"Bucket Name","placeholder":"Bucket Name","type":"SingleLineText","required":true},{"key":"region","label":"Region","placeholder":"Region","type":"SingleLineText","required":true},{"key":"access_key","label":"Access Key","placeholder":"Access Key","type":"SingleLineText","required":true},{"key":"access_secret","label":"Access Secret","placeholder":"Access Secret","type":"Password","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and attachment will be stored in OvhCloud Object Storage","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.674337+02	2022-05-18 11:22:25.674337+02
nc_e7c27x1bkofdkk	Linode Object Storage	S3-compatible Linode Object Storage makes it easy and more affordable to manage unstructured data such as content assets, as well as sophisticated and data-intensive storage challenges around artificial intelligence and machine learning.	f	\N	0.0.1	\N	install	\N	plugins/linode.svg	\N	Storage	Storage	{"title":"Configure Linode Object Storage","items":[{"key":"bucket","label":"Bucket Name","placeholder":"Bucket Name","type":"SingleLineText","required":true},{"key":"region","label":"Region","placeholder":"Region","type":"SingleLineText","required":true},{"key":"access_key","label":"Access Key","placeholder":"Access Key","type":"SingleLineText","required":true},{"key":"access_secret","label":"Access Secret","placeholder":"Access Secret","type":"Password","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and attachment will be stored in Linode Object Storage","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.675915+02	2022-05-18 11:22:25.675915+02
nc_k25p2hjlebeur9	UpCloud Object Storage	The perfect home for your data. Thanks to the S3-compatible programmable interface,\nyou have a host of options for existing tools and code implementations.\n	f	\N	0.0.1	\N	install	\N	plugins/upcloud.png	\N	Storage	Storage	{"title":"Configure UpCloud Object Storage","items":[{"key":"bucket","label":"Bucket Name","placeholder":"Bucket Name","type":"SingleLineText","required":true},{"key":"endpoint","label":"Endpoint","placeholder":"Endpoint","type":"SingleLineText","required":true},{"key":"access_key","label":"Access Key","placeholder":"Access Key","type":"SingleLineText","required":true},{"key":"access_secret","label":"Access Secret","placeholder":"Access Secret","type":"Password","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and attachment will be stored in UpCloud Object Storage","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.677397+02	2022-05-18 11:22:25.677397+02
nc_9eltn35kg5mq61	SMTP	SMTP email client	f	\N	0.0.1	\N	install	\N	\N	\N	Email	Email	{"title":"Configure Email SMTP","items":[{"key":"from","label":"From","placeholder":"eg: admin@example.com","type":"SingleLineText","required":true},{"key":"host","label":"Host","placeholder":"eg: smtp.example.com","type":"SingleLineText","required":true},{"key":"port","label":"Port","placeholder":"Port","type":"SingleLineText","required":true},{"key":"secure","label":"Secure","placeholder":"Secure","type":"SingleLineText","required":true},{"key":"ignoreTLS","label":"Ignore TLS","placeholder":"Ignore TLS","type":"Checkbox","required":false},{"key":"username","label":"Username","placeholder":"Username","type":"SingleLineText","required":false},{"key":"password","label":"Password","placeholder":"Password","type":"Password","required":false}],"actions":[{"label":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and email notification will use SMTP configuration","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.678897+02	2022-05-18 11:22:25.678897+02
nc_ui8edyp189nc7c	MailerSend	MailerSend email client	f	\N	0.0.1	\N	install	\N	plugins/mailersend.svg	\N	Email	Email	{"title":"Configure MailerSend","items":[{"key":"api_key","label":"API KEy","placeholder":"eg: ***************","type":"Password","required":true},{"key":"from","label":"From","placeholder":"eg: admin@example.com","type":"SingleLineText","required":true},{"key":"from_name","label":"From Name","placeholder":"eg: Adam","type":"SingleLineText","required":true}],"actions":[{"label":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and email notification will use MailerSend configuration","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.680327+02	2022-05-18 11:22:25.680327+02
nc_bcvyxg5bv6d97n	Scaleway Object Storage	Scaleway Object Storage is an S3-compatible object store from Scaleway Cloud Platform.	f	\N	0.0.1	\N	install	\N	plugins/scaleway.png	\N	Storage	Storage	{"title":"Setup Scaleway","items":[{"key":"bucket","label":"Bucket name","placeholder":"Bucket name","type":"SingleLineText","required":true},{"key":"region","label":"Region of bucket","placeholder":"Region of bucket","type":"SingleLineText","required":true},{"key":"access_key","label":"Access Key","placeholder":"Access Key","type":"SingleLineText","required":true},{"key":"access_secret","label":"Access Secret","placeholder":"Access Secret","type":"Password","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed Scaleway Object Storage","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.6818+02	2022-05-18 11:22:25.6818+02
nc_7iq75rma2zc6y6	SES	Amazon Simple Email Service (SES) is a cost-effective, flexible, and scalable email service that enables developers to send mail from within any application.	f	\N	0.0.1	\N	install	\N	plugins/aws.png	\N	Email	Email	{"title":"Configure Amazon Simple Email Service (SES)","items":[{"key":"from","label":"From","placeholder":"From","type":"SingleLineText","required":true},{"key":"region","label":"Region","placeholder":"Region","type":"SingleLineText","required":true},{"key":"access_key","label":"Access Key","placeholder":"Access Key","type":"SingleLineText","required":true},{"key":"access_secret","label":"Access Secret","placeholder":"Access Secret","type":"Password","required":true}],"actions":[{"label":"Test","placeholder":"Test","key":"test","actionType":"TEST","type":"Button"},{"label":"Save","placeholder":"Save","key":"save","actionType":"SUBMIT","type":"Button"}],"msgOnInstall":"Successfully installed and email notification will use Amazon SES","msgOnUninstall":""}	\N	\N	\N	\N	2022-05-18 11:22:25.683245+02	2022-05-18 11:22:25.683245+02
\.


--
-- Data for Name: nc_project_users_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_project_users_v2 (project_id, fk_user_id, roles, starred, pinned, "group", color, "order", hidden, opened_date, created_at, updated_at) FROM stdin;
p_41297240e6of48	us_j0y116g1uem0oi	owner	\N	\N	\N	\N	\N	\N	\N	2022-05-18 11:24:10.789828+02	2022-05-18 11:24:10.789828+02
\.


--
-- Data for Name: nc_projects; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_projects (id, title, status, description, config, meta, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_projects_users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_projects_users (project_id, user_id, roles, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_projects_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_projects_v2 (id, title, prefix, status, description, meta, color, uuid, password, roles, deleted, is_meta, "order", created_at, updated_at) FROM stdin;
p_41297240e6of48	northwind	\N	\N	\N	\N	\N	\N	\N	\N	f	f	\N	2022-05-18 11:24:10.7795+02	2022-05-18 11:24:10.7795+02
\.


--
-- Data for Name: nc_relations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_relations (id, project_id, db_alias, tn, rtn, _tn, _rtn, cn, rcn, _cn, _rcn, referenced_db_alias, type, db_type, ur, dr, created_at, updated_at, fkn) FROM stdin;
\.


--
-- Data for Name: nc_resolvers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_resolvers (id, project_id, db_alias, title, resolver, type, acl, functions, handler_type, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_roles; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_roles (id, project_id, db_alias, title, type, description, created_at, updated_at) FROM stdin;
1			owner	SYSTEM	Can add/remove creators. And full edit database structures & fields.	\N	\N
2			creator	SYSTEM	Can fully edit database structure & values	\N	\N
3			editor	SYSTEM	Can edit records but cannot change structure of database/fields	\N	\N
4			commenter	SYSTEM	Can view and comment the records but cannot edit anything	\N	\N
5			viewer	SYSTEM	Can view the records but cannot edit anything	\N	\N
\.


--
-- Data for Name: nc_routes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_routes (id, project_id, db_alias, title, tn, tnp, tnc, relation_type, path, type, handler, acl, "order", functions, handler_type, is_custom, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_rpc; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_rpc (id, project_id, db_alias, title, tn, service, tnp, tnc, relation_type, "order", type, acl, functions, handler_type, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_shared_bases; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_shared_bases (id, project_id, db_alias, roles, shared_base_id, enabled, password, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_shared_views; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_shared_views (id, project_id, db_alias, model_name, meta, query_params, view_id, show_all_fields, allow_copy, password, created_at, updated_at, view_type, view_name) FROM stdin;
\.


--
-- Data for Name: nc_shared_views_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_shared_views_v2 (id, fk_view_id, meta, query_params, view_id, show_all_fields, allow_copy, password, deleted, "order", created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_sort_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_sort_v2 (id, base_id, project_id, fk_view_id, fk_column_id, direction, "order", created_at, updated_at) FROM stdin;
so_phs90ir5tr5crg	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_ztmg51h4hma3hj	cl_x801ithwdkcrx9	asc	2	2022-05-18 11:27:49.585597+02	2022-05-18 11:27:49.585597+02
so_2v0wzoa5gmpypz	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_ztmg51h4hma3hj	cl_68x1go22132wqm	desc	1	2022-05-18 11:27:42.153102+02	2022-05-18 11:28:21.647413+02
so_j4izpi7tv68o1b	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_29tn3xki80n7e3	cl_42mxp318j7y76i	asc	1	2022-05-18 11:50:22.867494+02	2022-05-18 11:50:22.867494+02
so_nu0nunya0j4rx9	ds_aej9ao4mzw9cff	p_41297240e6of48	vw_qkcvmkknuwc3le	cl_42mxp318j7y76i	asc	1	2022-05-18 12:02:23.17372+02	2022-05-18 12:02:23.17372+02
\.


--
-- Data for Name: nc_store; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_store (id, project_id, db_alias, key, value, type, env, tag, created_at, updated_at) FROM stdin;
1	\N		NC_DEBUG	{"nc:app":false,"nc:api:rest":false,"nc:api:base":false,"nc:api:gql":false,"nc:api:grpc":false,"nc:migrator":false,"nc:datamapper":false}	\N	\N	\N	\N	\N
2	\N		NC_PROJECT_COUNT	0	\N	\N	\N	\N	\N
3			nc_auth_jwt_secret	678bd776-b8b4-414f-9d79-3f3c73608b14	\N	\N	\N	2022-05-18 11:22:25.552869+02	2022-05-18 11:22:25.552869+02
4			nc_server_id	188e202c602acf0ea86242985475e6abb7326b381fb04107bae5fbe3a5182ff2	\N	\N	\N	2022-05-18 11:22:25.56282+02	2022-05-18 11:22:25.56282+02
5			NC_CONFIG_MAIN	{"version":"0090000"}	\N	\N	\N	2022-05-18 11:22:25.564162+02	2022-05-18 11:22:25.564162+02
\.


--
-- Data for Name: nc_team_users_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_team_users_v2 (org_id, user_id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_teams_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_teams_v2 (id, title, org_id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: nc_users_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_users_v2 (id, email, password, salt, firstname, lastname, username, refresh_token, invite_token, invite_token_expires, reset_password_expires, reset_password_token, email_verification_token, email_verified, roles, created_at, updated_at) FROM stdin;
us_j0y116g1uem0oi	postgreserikjonsson@yahoo.fr	$2a$10$xTtSh0R1i8FK9H2iTI7uruuvEDOp0B7.mcGnbkqrTan.paygx87Ey	$2a$10$xTtSh0R1i8FK9H2iTI7uru	\N	\N	\N	eafb6ec0df89e0231fd61757cd0fe8e285814392db3f502b5b44dab50196f1efee39913adb5b1635	\N	\N	\N	\N	c7575841-bdf4-410a-8959-cc921fbbc48d	\N	user,super	2022-05-18 11:23:15.906752+02	2022-05-18 11:23:15.91513+02
\.


--
-- Data for Name: nc_views_v2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nc_views_v2 (id, base_id, project_id, fk_model_id, title, type, is_default, show_system_fields, lock_type, uuid, password, show, "order", created_at, updated_at) FROM stdin;
vw_om33x8ttx67est	ds_aej9ao4mzw9cff	p_41297240e6of48	md_79w8haiknxbkzd	Categories	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.635273+02	2022-05-18 11:24:11.635273+02
vw_xfg9ot5ntsq77q	ds_aej9ao4mzw9cff	p_41297240e6of48	md_7vpsxklk3sph75	CustomerCustomerDemo	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.640145+02	2022-05-18 11:24:11.640145+02
vw_c525afyrdh8nqn	ds_aej9ao4mzw9cff	p_41297240e6of48	md_71j5r6adaq9l3i	CustomerDemographics	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.640881+02	2022-05-18 11:24:11.640881+02
vw_tsu33qayws91ma	ds_aej9ao4mzw9cff	p_41297240e6of48	md_vb4nqekkid2rl6	Customers	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.660823+02	2022-05-18 11:24:11.660823+02
vw_mxy2on3nt3kj0g	ds_aej9ao4mzw9cff	p_41297240e6of48	md_q2z56jvps22ukx	EmployeeTerritories	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.665016+02	2022-05-18 11:24:11.665016+02
vw_jq0zcol66s3voo	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yi49xujy865heu	NcBasesV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.680902+02	2022-05-18 11:24:11.680902+02
vw_ipcsobcohj1b6c	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5hjnvbjhgfiz6r	NcColLookupV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.681709+02	2022-05-18 11:24:11.681709+02
vw_gl8cwxr6m3iq38	ds_aej9ao4mzw9cff	p_41297240e6of48	md_tm1aur9f3yn2l2	NcColFormulaV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.68269+02	2022-05-18 11:24:11.68269+02
vw_h61l96me3nfmwj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ue8cshscw72tz4	NcColumnsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.683861+02	2022-05-18 11:24:11.683861+02
vw_u2jzjq9smernol	ds_aej9ao4mzw9cff	p_41297240e6of48	md_yor91fufon7h6c	Employees	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.684328+02	2022-05-18 11:24:11.684328+02
vw_2lgkfm9g3xygp1	ds_aej9ao4mzw9cff	p_41297240e6of48	md_91fcdvmtveuz77	NcDisabledModelsForRoleV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.684839+02	2022-05-18 11:24:11.684839+02
vw_ewjjjsi3mbc5h9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ruvc4yz1y8yonz	NcColRelationsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.685242+02	2022-05-18 11:24:11.685242+02
vw_dzgaxceud97gqt	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6pjpqivc2ymh3b	NcColSelectOptionsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.685637+02	2022-05-18 11:24:11.685637+02
vw_jbveqkpzj241da	ds_aej9ao4mzw9cff	p_41297240e6of48	md_hle53jhinff293	NcAuditV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.686129+02	2022-05-18 11:24:11.686129+02
vw_sxajywo4jx9ml5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_rmg6p4gh772ip8	NcColRollupV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.687286+02	2022-05-18 11:24:11.687286+02
vw_hk5zatfifol9yl	ds_aej9ao4mzw9cff	p_41297240e6of48	md_gwspodfugzyzs2	NcFilterExpV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.687933+02	2022-05-18 11:24:11.687933+02
vw_pcdpz0lqr00s9k	ds_aej9ao4mzw9cff	p_41297240e6of48	md_op8dq4us6zg5tn	NcFormViewColumnsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.688971+02	2022-05-18 11:24:11.688971+02
vw_etd91poe0pvyld	ds_aej9ao4mzw9cff	p_41297240e6of48	md_47c5g91l1nc5se	NcFormViewV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.693776+02	2022-05-18 11:24:11.693776+02
vw_n362p00huzpsod	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ajkugjxa4haf22	NcKanbanViewV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.696078+02	2022-05-18 11:24:11.696078+02
vw_wflzr4fus2ngby	ds_aej9ao4mzw9cff	p_41297240e6of48	md_thf0xauk4e6qxs	NcGridViewColumnsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.698864+02	2022-05-18 11:24:11.698864+02
vw_andyomdmvq8tv6	ds_aej9ao4mzw9cff	p_41297240e6of48	md_2vqez05sn215dj	NcGalleryViewV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.699331+02	2022-05-18 11:24:11.699331+02
vw_7fueohglirse69	ds_aej9ao4mzw9cff	p_41297240e6of48	md_zmhiordb5vpwkf	NcGridViewV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.699861+02	2022-05-18 11:24:11.699861+02
vw_go1u33h41idruk	ds_aej9ao4mzw9cff	p_41297240e6of48	md_whwvfz6cb3ljdb	NcGalleryViewColumnsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.700457+02	2022-05-18 11:24:11.700457+02
vw_avtdalhk6frgn9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mqny0xchyq3gqp	NcHookLogsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.701011+02	2022-05-18 11:24:11.701011+02
vw_ea8mjggb90pnnt	ds_aej9ao4mzw9cff	p_41297240e6of48	md_muq975426eve8k	NcHooksV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.701318+02	2022-05-18 11:24:11.701318+02
vw_6exoiii01n21je	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ejwlk09y393nhf	NcKanbanViewColumnsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.704052+02	2022-05-18 11:24:11.704052+02
vw_352nd6p9zohj5w	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ebe9bibmq1l2uf	NcProjectUsersV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.707293+02	2022-05-18 11:24:11.707293+02
vw_lalt4t2bjas0gj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_5ms9hq3t1a9cro	NcPluginsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.709014+02	2022-05-18 11:24:11.709014+02
vw_bg9irqfa68gtl3	ds_aej9ao4mzw9cff	p_41297240e6of48	md_ofa8d1ocm4oowx	NcProjectsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.71142+02	2022-05-18 11:24:11.71142+02
vw_q1cl9wv3hhuvca	ds_aej9ao4mzw9cff	p_41297240e6of48	md_nj2q6tfni3b5a8	NcOrgsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.712516+02	2022-05-18 11:24:11.712516+02
vw_brn7rul3d130ex	ds_aej9ao4mzw9cff	p_41297240e6of48	md_1wgrn82853mqsm	NcSharedBases	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.713594+02	2022-05-18 11:24:11.713594+02
vw_lr57pe50bbmjar	ds_aej9ao4mzw9cff	p_41297240e6of48	md_0z5ofu8o9xph1j	NcTeamsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.714313+02	2022-05-18 11:24:11.714313+02
vw_7ppnzs3y6eip9i	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wqy1a8c5ggpyea	NcSortV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.714826+02	2022-05-18 11:24:11.714826+02
vw_mauaaqk5b7v0rr	ds_aej9ao4mzw9cff	p_41297240e6of48	md_12azccujctejop	NcSharedViewsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.715277+02	2022-05-18 11:24:11.715277+02
vw_51d0c7bmbclwb8	ds_aej9ao4mzw9cff	p_41297240e6of48	md_89v98o1rpi4idv	NcTeamUsersV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.717057+02	2022-05-18 11:24:11.717057+02
vw_t8o4jwxtrx9nf0	ds_aej9ao4mzw9cff	p_41297240e6of48	md_cltahxtbesn8vs	NcModelsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.719231+02	2022-05-18 11:24:11.719231+02
vw_dejssdomuxg3uw	ds_aej9ao4mzw9cff	p_41297240e6of48	md_6ht8y7vgzcek0p	NcUsersV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.721873+02	2022-05-18 11:24:11.721873+02
vw_29tn3xki80n7e3	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	Orders	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.731822+02	2022-05-18 11:24:11.731822+02
vw_ztmg51h4hma3hj	ds_aej9ao4mzw9cff	p_41297240e6of48	md_mmp2o2atqtudsn	Products	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.73443+02	2022-05-18 11:24:11.73443+02
vw_du7zrxhj51z27u	ds_aej9ao4mzw9cff	p_41297240e6of48	md_jgq6gbzvyfbk2g	Region	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.734895+02	2022-05-18 11:24:11.734895+02
vw_ccggsien3mx2nl	ds_aej9ao4mzw9cff	p_41297240e6of48	md_s9jny9njrhdnwz	OrderDetails	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.735736+02	2022-05-18 11:24:11.735736+02
vw_5w6mg5ux0ryjjl	ds_aej9ao4mzw9cff	p_41297240e6of48	md_eqrk8yd76gurbf	Suppliers	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.736885+02	2022-05-18 11:24:11.736885+02
vw_r15acq81tqlkxp	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wk0cg5v1lnk9ew	UsStates	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.737144+02	2022-05-18 11:24:11.737144+02
vw_srnlvrf0gn6rul	ds_aej9ao4mzw9cff	p_41297240e6of48	md_dg3rmvfk4bsg1i	NcViewsV2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.737502+02	2022-05-18 11:24:11.737502+02
vw_81imke2v3e15l5	ds_aej9ao4mzw9cff	p_41297240e6of48	md_j409llgiqo90yz	Territories	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.742171+02	2022-05-18 11:24:11.742171+02
vw_idcyonon9t2t5l	ds_aej9ao4mzw9cff	p_41297240e6of48	md_wjrbh5du4ifq5b	Shippers	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.743052+02	2022-05-18 11:24:11.743052+02
vw_yrmq5jknxj9sx6	ds_aej9ao4mzw9cff	p_41297240e6of48	md_llr0f8esceyorr	XcKnexMigrationsv2	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.743335+02	2022-05-18 11:24:11.743335+02
vw_8kwir8e38t3fc9	ds_aej9ao4mzw9cff	p_41297240e6of48	md_c1y2qwn3okt75w	XcKnexMigrationsv2Lock	3	t	\N	collaborative	\N	\N	t	1	2022-05-18 11:24:11.752707+02	2022-05-18 11:24:11.752707+02
vw_qkcvmkknuwc3le	ds_aej9ao4mzw9cff	p_41297240e6of48	md_r7cgbvzh9qy7qk	Orders1	1	\N	\N	collaborative	\N	\N	t	2	2022-05-18 12:02:23.162955+02	2022-05-18 12:02:23.162955+02
\.


--
-- Data for Name: order_details; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.order_details (order_id, product_id, unit_price, quantity, discount) FROM stdin;
10248	11	14	12	0
10248	42	9.8	10	0
10248	72	34.8	5	0
10249	14	18.6	9	0
10249	51	42.4	40	0
10250	41	7.7	10	0
10250	51	42.4	35	0.15
10250	65	16.8	15	0.15
10251	22	16.8	6	0.05
10251	57	15.6	15	0.05
10251	65	16.8	20	0
10252	20	64.8	40	0.05
10252	33	2	25	0.05
10252	60	27.2	40	0
10253	31	10	20	0
10253	39	14.4	42	0
10253	49	16	40	0
10254	24	3.6	15	0.15
10254	55	19.2	21	0.15
10254	74	8	21	0
10255	2	15.2	20	0
10255	16	13.9	35	0
10255	36	15.2	25	0
10255	59	44	30	0
10256	53	26.2	15	0
10256	77	10.4	12	0
10257	27	35.1	25	0
10257	39	14.4	6	0
10257	77	10.4	15	0
10258	2	15.2	50	0.2
10258	5	17	65	0.2
10258	32	25.6	6	0.2
10259	21	8	10	0
10259	37	20.8	1	0
10260	41	7.7	16	0.25
10260	57	15.6	50	0
10260	62	39.4	15	0.25
10260	70	12	21	0.25
10261	21	8	20	0
10261	35	14.4	20	0
10262	5	17	12	0.2
10262	7	24	15	0
10262	56	30.4	2	0
10263	16	13.9	60	0.25
10263	24	3.6	28	0
10263	30	20.7	60	0.25
10263	74	8	36	0.25
10264	2	15.2	35	0
10264	41	7.7	25	0.15
10265	17	31.2	30	0
10265	70	12	20	0
10266	12	30.4	12	0.05
10267	40	14.7	50	0
10267	59	44	70	0.15
10267	76	14.4	15	0.15
10268	29	99	10	0
10268	72	27.8	4	0
10269	33	2	60	0.05
10269	72	27.8	20	0.05
10270	36	15.2	30	0
10270	43	36.8	25	0
10271	33	2	24	0
10272	20	64.8	6	0
10272	31	10	40	0
10272	72	27.8	24	0
10273	10	24.8	24	0.05
10273	31	10	15	0.05
10273	33	2	20	0
10273	40	14.7	60	0.05
10273	76	14.4	33	0.05
10274	71	17.2	20	0
10274	72	27.8	7	0
10275	24	3.6	12	0.05
10275	59	44	6	0.05
10276	10	24.8	15	0
10276	13	4.8	10	0
10277	28	36.4	20	0
10277	62	39.4	12	0
10278	44	15.5	16	0
10278	59	44	15	0
10278	63	35.1	8	0
10278	73	12	25	0
10279	17	31.2	15	0.25
10280	24	3.6	12	0
10280	55	19.2	20	0
10280	75	6.2	30	0
10281	19	7.3	1	0
10281	24	3.6	6	0
10281	35	14.4	4	0
10282	30	20.7	6	0
10282	57	15.6	2	0
10283	15	12.4	20	0
10283	19	7.3	18	0
10283	60	27.2	35	0
10283	72	27.8	3	0
10284	27	35.1	15	0.25
10284	44	15.5	21	0
10284	60	27.2	20	0.25
10284	67	11.2	5	0.25
10285	1	14.4	45	0.2
10285	40	14.7	40	0.2
10285	53	26.2	36	0.2
10286	35	14.4	100	0
10286	62	39.4	40	0
10287	16	13.9	40	0.15
10287	34	11.2	20	0
10287	46	9.6	15	0.15
10288	54	5.9	10	0.1
10288	68	10	3	0.1
10289	3	8	30	0
10289	64	26.6	9	0
10290	5	17	20	0
10290	29	99	15	0
10290	49	16	15	0
10290	77	10.4	10	0
10291	13	4.8	20	0.1
10291	44	15.5	24	0.1
10291	51	42.4	2	0.1
10292	20	64.8	20	0
10293	18	50	12	0
10293	24	3.6	10	0
10293	63	35.1	5	0
10293	75	6.2	6	0
10294	1	14.4	18	0
10294	17	31.2	15	0
10294	43	36.8	15	0
10294	60	27.2	21	0
10294	75	6.2	6	0
10295	56	30.4	4	0
10296	11	16.8	12	0
10296	16	13.9	30	0
10296	69	28.8	15	0
10297	39	14.4	60	0
10297	72	27.8	20	0
10298	2	15.2	40	0
10298	36	15.2	40	0.25
10298	59	44	30	0.25
10298	62	39.4	15	0
10299	19	7.3	15	0
10299	70	12	20	0
10300	66	13.6	30	0
10300	68	10	20	0
10301	40	14.7	10	0
10301	56	30.4	20	0
10302	17	31.2	40	0
10302	28	36.4	28	0
10302	43	36.8	12	0
10303	40	14.7	40	0.1
10303	65	16.8	30	0.1
10303	68	10	15	0.1
10304	49	16	30	0
10304	59	44	10	0
10304	71	17.2	2	0
10305	18	50	25	0.1
10305	29	99	25	0.1
10305	39	14.4	30	0.1
10306	30	20.7	10	0
10306	53	26.2	10	0
10306	54	5.9	5	0
10307	62	39.4	10	0
10307	68	10	3	0
10308	69	28.8	1	0
10308	70	12	5	0
10309	4	17.6	20	0
10309	6	20	30	0
10309	42	11.2	2	0
10309	43	36.8	20	0
10309	71	17.2	3	0
10310	16	13.9	10	0
10310	62	39.4	5	0
10311	42	11.2	6	0
10311	69	28.8	7	0
10312	28	36.4	4	0
10312	43	36.8	24	0
10312	53	26.2	20	0
10312	75	6.2	10	0
10313	36	15.2	12	0
10314	32	25.6	40	0.1
10314	58	10.6	30	0.1
10314	62	39.4	25	0.1
10315	34	11.2	14	0
10315	70	12	30	0
10316	41	7.7	10	0
10316	62	39.4	70	0
10317	1	14.4	20	0
10318	41	7.7	20	0
10318	76	14.4	6	0
10319	17	31.2	8	0
10319	28	36.4	14	0
10319	76	14.4	30	0
10320	71	17.2	30	0
10321	35	14.4	10	0
10322	52	5.6	20	0
10323	15	12.4	5	0
10323	25	11.2	4	0
10323	39	14.4	4	0
10324	16	13.9	21	0.15
10324	35	14.4	70	0.15
10324	46	9.6	30	0
10324	59	44	40	0.15
10324	63	35.1	80	0.15
10325	6	20	6	0
10325	13	4.8	12	0
10325	14	18.6	9	0
10325	31	10	4	0
10325	72	27.8	40	0
10326	4	17.6	24	0
10326	57	15.6	16	0
10326	75	6.2	50	0
10327	2	15.2	25	0.2
10327	11	16.8	50	0.2
10327	30	20.7	35	0.2
10327	58	10.6	30	0.2
10328	59	44	9	0
10328	65	16.8	40	0
10328	68	10	10	0
10329	19	7.3	10	0.05
10329	30	20.7	8	0.05
10329	38	210.8	20	0.05
10329	56	30.4	12	0.05
10330	26	24.9	50	0.15
10330	72	27.8	25	0.15
10331	54	5.9	15	0
10332	18	50	40	0.2
10332	42	11.2	10	0.2
10332	47	7.6	16	0.2
10333	14	18.6	10	0
10333	21	8	10	0.1
10333	71	17.2	40	0.1
10334	52	5.6	8	0
10334	68	10	10	0
10335	2	15.2	7	0.2
10335	31	10	25	0.2
10335	32	25.6	6	0.2
10335	51	42.4	48	0.2
10336	4	17.6	18	0.1
10337	23	7.2	40	0
10337	26	24.9	24	0
10337	36	15.2	20	0
10337	37	20.8	28	0
10337	72	27.8	25	0
10338	17	31.2	20	0
10338	30	20.7	15	0
10339	4	17.6	10	0
10339	17	31.2	70	0.05
10339	62	39.4	28	0
10340	18	50	20	0.05
10340	41	7.7	12	0.05
10340	43	36.8	40	0.05
10341	33	2	8	0
10341	59	44	9	0.15
10342	2	15.2	24	0.2
10342	31	10	56	0.2
10342	36	15.2	40	0.2
10342	55	19.2	40	0.2
10343	64	26.6	50	0
10343	68	10	4	0.05
10343	76	14.4	15	0
10344	4	17.6	35	0
10344	8	32	70	0.25
10345	8	32	70	0
10345	19	7.3	80	0
10345	42	11.2	9	0
10346	17	31.2	36	0.1
10346	56	30.4	20	0
10347	25	11.2	10	0
10347	39	14.4	50	0.15
10347	40	14.7	4	0
10347	75	6.2	6	0.15
10348	1	14.4	15	0.15
10348	23	7.2	25	0
10349	54	5.9	24	0
10350	50	13	15	0.1
10350	69	28.8	18	0.1
10351	38	210.8	20	0.05
10351	41	7.7	13	0
10351	44	15.5	77	0.05
10351	65	16.8	10	0.05
10352	24	3.6	10	0
10352	54	5.9	20	0.15
10353	11	16.8	12	0.2
10353	38	210.8	50	0.2
10354	1	14.4	12	0
10354	29	99	4	0
10355	24	3.6	25	0
10355	57	15.6	25	0
10356	31	10	30	0
10356	55	19.2	12	0
10356	69	28.8	20	0
10357	10	24.8	30	0.2
10357	26	24.9	16	0
10357	60	27.2	8	0.2
10358	24	3.6	10	0.05
10358	34	11.2	10	0.05
10358	36	15.2	20	0.05
10359	16	13.9	56	0.05
10359	31	10	70	0.05
10359	60	27.2	80	0.05
10360	28	36.4	30	0
10360	29	99	35	0
10360	38	210.8	10	0
10360	49	16	35	0
10360	54	5.9	28	0
10361	39	14.4	54	0.1
10361	60	27.2	55	0.1
10362	25	11.2	50	0
10362	51	42.4	20	0
10362	54	5.9	24	0
10363	31	10	20	0
10363	75	6.2	12	0
10363	76	14.4	12	0
10364	69	28.8	30	0
10364	71	17.2	5	0
10365	11	16.8	24	0
10366	65	16.8	5	0
10366	77	10.4	5	0
10367	34	11.2	36	0
10367	54	5.9	18	0
10367	65	16.8	15	0
10367	77	10.4	7	0
10368	21	8	5	0.1
10368	28	36.4	13	0.1
10368	57	15.6	25	0
10368	64	26.6	35	0.1
10369	29	99	20	0
10369	56	30.4	18	0.25
10370	1	14.4	15	0.15
10370	64	26.6	30	0
10370	74	8	20	0.15
10371	36	15.2	6	0.2
10372	20	64.8	12	0.25
10372	38	210.8	40	0.25
10372	60	27.2	70	0.25
10372	72	27.8	42	0.25
10373	58	10.6	80	0.2
10373	71	17.2	50	0.2
10374	31	10	30	0
10374	58	10.6	15	0
10375	14	18.6	15	0
10375	54	5.9	10	0
10376	31	10	42	0.05
10377	28	36.4	20	0.15
10377	39	14.4	20	0.15
10378	71	17.2	6	0
10379	41	7.7	8	0.1
10379	63	35.1	16	0.1
10379	65	16.8	20	0.1
10380	30	20.7	18	0.1
10380	53	26.2	20	0.1
10380	60	27.2	6	0.1
10380	70	12	30	0
10381	74	8	14	0
10382	5	17	32	0
10382	18	50	9	0
10382	29	99	14	0
10382	33	2	60	0
10382	74	8	50	0
10383	13	4.8	20	0
10383	50	13	15	0
10383	56	30.4	20	0
10384	20	64.8	28	0
10384	60	27.2	15	0
10385	7	24	10	0.2
10385	60	27.2	20	0.2
10385	68	10	8	0.2
10386	24	3.6	15	0
10386	34	11.2	10	0
10387	24	3.6	15	0
10387	28	36.4	6	0
10387	59	44	12	0
10387	71	17.2	15	0
10388	45	7.6	15	0.2
10388	52	5.6	20	0.2
10388	53	26.2	40	0
10389	10	24.8	16	0
10389	55	19.2	15	0
10389	62	39.4	20	0
10389	70	12	30	0
10390	31	10	60	0.1
10390	35	14.4	40	0.1
10390	46	9.6	45	0
10390	72	27.8	24	0.1
10391	13	4.8	18	0
10392	69	28.8	50	0
10393	2	15.2	25	0.25
10393	14	18.6	42	0.25
10393	25	11.2	7	0.25
10393	26	24.9	70	0.25
10393	31	10	32	0
10394	13	4.8	10	0
10394	62	39.4	10	0
10395	46	9.6	28	0.1
10395	53	26.2	70	0.1
10395	69	28.8	8	0
10396	23	7.2	40	0
10396	71	17.2	60	0
10396	72	27.8	21	0
10397	21	8	10	0.15
10397	51	42.4	18	0.15
10398	35	14.4	30	0
10398	55	19.2	120	0.1
10399	68	10	60	0
10399	71	17.2	30	0
10399	76	14.4	35	0
10399	77	10.4	14	0
10400	29	99	21	0
10400	35	14.4	35	0
10400	49	16	30	0
10401	30	20.7	18	0
10401	56	30.4	70	0
10401	65	16.8	20	0
10401	71	17.2	60	0
10402	23	7.2	60	0
10402	63	35.1	65	0
10403	16	13.9	21	0.15
10403	48	10.2	70	0.15
10404	26	24.9	30	0.05
10404	42	11.2	40	0.05
10404	49	16	30	0.05
10405	3	8	50	0
10406	1	14.4	10	0
10406	21	8	30	0.1
10406	28	36.4	42	0.1
10406	36	15.2	5	0.1
10406	40	14.7	2	0.1
10407	11	16.8	30	0
10407	69	28.8	15	0
10407	71	17.2	15	0
10408	37	20.8	10	0
10408	54	5.9	6	0
10408	62	39.4	35	0
10409	14	18.6	12	0
10409	21	8	12	0
10410	33	2	49	0
10410	59	44	16	0
10411	41	7.7	25	0.2
10411	44	15.5	40	0.2
10411	59	44	9	0.2
10412	14	18.6	20	0.1
10413	1	14.4	24	0
10413	62	39.4	40	0
10413	76	14.4	14	0
10414	19	7.3	18	0.05
10414	33	2	50	0
10415	17	31.2	2	0
10415	33	2	20	0
10416	19	7.3	20	0
10416	53	26.2	10	0
10416	57	15.6	20	0
10417	38	210.8	50	0
10417	46	9.6	2	0.25
10417	68	10	36	0.25
10417	77	10.4	35	0
10418	2	15.2	60	0
10418	47	7.6	55	0
10418	61	22.8	16	0
10418	74	8	15	0
10419	60	27.2	60	0.05
10419	69	28.8	20	0.05
10420	9	77.6	20	0.1
10420	13	4.8	2	0.1
10420	70	12	8	0.1
10420	73	12	20	0.1
10421	19	7.3	4	0.15
10421	26	24.9	30	0
10421	53	26.2	15	0.15
10421	77	10.4	10	0.15
10422	26	24.9	2	0
10423	31	10	14	0
10423	59	44	20	0
10424	35	14.4	60	0.2
10424	38	210.8	49	0.2
10424	68	10	30	0.2
10425	55	19.2	10	0.25
10425	76	14.4	20	0.25
10426	56	30.4	5	0
10426	64	26.6	7	0
10427	14	18.6	35	0
10428	46	9.6	20	0
10429	50	13	40	0
10429	63	35.1	35	0.25
10430	17	31.2	45	0.2
10430	21	8	50	0
10430	56	30.4	30	0
10430	59	44	70	0.2
10431	17	31.2	50	0.25
10431	40	14.7	50	0.25
10431	47	7.6	30	0.25
10432	26	24.9	10	0
10432	54	5.9	40	0
10433	56	30.4	28	0
10434	11	16.8	6	0
10434	76	14.4	18	0.15
10435	2	15.2	10	0
10435	22	16.8	12	0
10435	72	27.8	10	0
10436	46	9.6	5	0
10436	56	30.4	40	0.1
10436	64	26.6	30	0.1
10436	75	6.2	24	0.1
10437	53	26.2	15	0
10438	19	7.3	15	0.2
10438	34	11.2	20	0.2
10438	57	15.6	15	0.2
10439	12	30.4	15	0
10439	16	13.9	16	0
10439	64	26.6	6	0
10439	74	8	30	0
10440	2	15.2	45	0.15
10440	16	13.9	49	0.15
10440	29	99	24	0.15
10440	61	22.8	90	0.15
10441	27	35.1	50	0
10442	11	16.8	30	0
10442	54	5.9	80	0
10442	66	13.6	60	0
10443	11	16.8	6	0.2
10443	28	36.4	12	0
10444	17	31.2	10	0
10444	26	24.9	15	0
10444	35	14.4	8	0
10444	41	7.7	30	0
10445	39	14.4	6	0
10445	54	5.9	15	0
10446	19	7.3	12	0.1
10446	24	3.6	20	0.1
10446	31	10	3	0.1
10446	52	5.6	15	0.1
10447	19	7.3	40	0
10447	65	16.8	35	0
10447	71	17.2	2	0
10448	26	24.9	6	0
10448	40	14.7	20	0
10449	10	24.8	14	0
10449	52	5.6	20	0
10449	62	39.4	35	0
10450	10	24.8	20	0.2
10450	54	5.9	6	0.2
10451	55	19.2	120	0.1
10451	64	26.6	35	0.1
10451	65	16.8	28	0.1
10451	77	10.4	55	0.1
10452	28	36.4	15	0
10452	44	15.5	100	0.05
10453	48	10.2	15	0.1
10453	70	12	25	0.1
10454	16	13.9	20	0.2
10454	33	2	20	0.2
10454	46	9.6	10	0.2
10455	39	14.4	20	0
10455	53	26.2	50	0
10455	61	22.8	25	0
10455	71	17.2	30	0
10456	21	8	40	0.15
10456	49	16	21	0.15
10457	59	44	36	0
10458	26	24.9	30	0
10458	28	36.4	30	0
10458	43	36.8	20	0
10458	56	30.4	15	0
10458	71	17.2	50	0
10459	7	24	16	0.05
10459	46	9.6	20	0.05
10459	72	27.8	40	0
10460	68	10	21	0.25
10460	75	6.2	4	0.25
10461	21	8	40	0.25
10461	30	20.7	28	0.25
10461	55	19.2	60	0.25
10462	13	4.8	1	0
10462	23	7.2	21	0
10463	19	7.3	21	0
10463	42	11.2	50	0
10464	4	17.6	16	0.2
10464	43	36.8	3	0
10464	56	30.4	30	0.2
10464	60	27.2	20	0
10465	24	3.6	25	0
10465	29	99	18	0.1
10465	40	14.7	20	0
10465	45	7.6	30	0.1
10465	50	13	25	0
10466	11	16.8	10	0
10466	46	9.6	5	0
10467	24	3.6	28	0
10467	25	11.2	12	0
10468	30	20.7	8	0
10468	43	36.8	15	0
10469	2	15.2	40	0.15
10469	16	13.9	35	0.15
10469	44	15.5	2	0.15
10470	18	50	30	0
10470	23	7.2	15	0
10470	64	26.6	8	0
10471	7	24	30	0
10471	56	30.4	20	0
10472	24	3.6	80	0.05
10472	51	42.4	18	0
10473	33	2	12	0
10473	71	17.2	12	0
10474	14	18.6	12	0
10474	28	36.4	18	0
10474	40	14.7	21	0
10474	75	6.2	10	0
10475	31	10	35	0.15
10475	66	13.6	60	0.15
10475	76	14.4	42	0.15
10476	55	19.2	2	0.05
10476	70	12	12	0
10477	1	14.4	15	0
10477	21	8	21	0.25
10477	39	14.4	20	0.25
10478	10	24.8	20	0.05
10479	38	210.8	30	0
10479	53	26.2	28	0
10479	59	44	60	0
10479	64	26.6	30	0
10480	47	7.6	30	0
10480	59	44	12	0
10481	49	16	24	0
10481	60	27.2	40	0
10482	40	14.7	10	0
10483	34	11.2	35	0.05
10483	77	10.4	30	0.05
10484	21	8	14	0
10484	40	14.7	10	0
10484	51	42.4	3	0
10485	2	15.2	20	0.1
10485	3	8	20	0.1
10485	55	19.2	30	0.1
10485	70	12	60	0.1
10486	11	16.8	5	0
10486	51	42.4	25	0
10486	74	8	16	0
10487	19	7.3	5	0
10487	26	24.9	30	0
10487	54	5.9	24	0.25
10488	59	44	30	0
10488	73	12	20	0.2
10489	11	16.8	15	0.25
10489	16	13.9	18	0
10490	59	44	60	0
10490	68	10	30	0
10490	75	6.2	36	0
10491	44	15.5	15	0.15
10491	77	10.4	7	0.15
10492	25	11.2	60	0.05
10492	42	11.2	20	0.05
10493	65	16.8	15	0.1
10493	66	13.6	10	0.1
10493	69	28.8	10	0.1
10494	56	30.4	30	0
10495	23	7.2	10	0
10495	41	7.7	20	0
10495	77	10.4	5	0
10496	31	10	20	0.05
10497	56	30.4	14	0
10497	72	27.8	25	0
10497	77	10.4	25	0
10498	24	4.5	14	0
10498	40	18.4	5	0
10498	42	14	30	0
10499	28	45.6	20	0
10499	49	20	25	0
10500	15	15.5	12	0.05
10500	28	45.6	8	0.05
10501	54	7.45	20	0
10502	45	9.5	21	0
10502	53	32.8	6	0
10502	67	14	30	0
10503	14	23.25	70	0
10503	65	21.05	20	0
10504	2	19	12	0
10504	21	10	12	0
10504	53	32.8	10	0
10504	61	28.5	25	0
10505	62	49.3	3	0
10506	25	14	18	0.1
10506	70	15	14	0.1
10507	43	46	15	0.15
10507	48	12.75	15	0.15
10508	13	6	10	0
10508	39	18	10	0
10509	28	45.6	3	0
10510	29	123.79	36	0
10510	75	7.75	36	0.1
10511	4	22	50	0.15
10511	7	30	50	0.15
10511	8	40	10	0.15
10512	24	4.5	10	0.15
10512	46	12	9	0.15
10512	47	9.5	6	0.15
10512	60	34	12	0.15
10513	21	10	40	0.2
10513	32	32	50	0.2
10513	61	28.5	15	0.2
10514	20	81	39	0
10514	28	45.6	35	0
10514	56	38	70	0
10514	65	21.05	39	0
10514	75	7.75	50	0
10515	9	97	16	0.15
10515	16	17.45	50	0
10515	27	43.9	120	0
10515	33	2.5	16	0.15
10515	60	34	84	0.15
10516	18	62.5	25	0.1
10516	41	9.65	80	0.1
10516	42	14	20	0
10517	52	7	6	0
10517	59	55	4	0
10517	70	15	6	0
10518	24	4.5	5	0
10518	38	263.5	15	0
10518	44	19.45	9	0
10519	10	31	16	0.05
10519	56	38	40	0
10519	60	34	10	0.05
10520	24	4.5	8	0
10520	53	32.8	5	0
10521	35	18	3	0
10521	41	9.65	10	0
10521	68	12.5	6	0
10522	1	18	40	0.2
10522	8	40	24	0
10522	30	25.89	20	0.2
10522	40	18.4	25	0.2
10523	17	39	25	0.1
10523	20	81	15	0.1
10523	37	26	18	0.1
10523	41	9.65	6	0.1
10524	10	31	2	0
10524	30	25.89	10	0
10524	43	46	60	0
10524	54	7.45	15	0
10525	36	19	30	0
10525	40	18.4	15	0.1
10526	1	18	8	0.15
10526	13	6	10	0
10526	56	38	30	0.15
10527	4	22	50	0.1
10527	36	19	30	0.1
10528	11	21	3	0
10528	33	2.5	8	0.2
10528	72	34.8	9	0
10529	55	24	14	0
10529	68	12.5	20	0
10529	69	36	10	0
10530	17	39	40	0
10530	43	46	25	0
10530	61	28.5	20	0
10530	76	18	50	0
10531	59	55	2	0
10532	30	25.89	15	0
10532	66	17	24	0
10533	4	22	50	0.05
10533	72	34.8	24	0
10533	73	15	24	0.05
10534	30	25.89	10	0
10534	40	18.4	10	0.2
10534	54	7.45	10	0.2
10535	11	21	50	0.1
10535	40	18.4	10	0.1
10535	57	19.5	5	0.1
10535	59	55	15	0.1
10536	12	38	15	0.25
10536	31	12.5	20	0
10536	33	2.5	30	0
10536	60	34	35	0.25
10537	31	12.5	30	0
10537	51	53	6	0
10537	58	13.25	20	0
10537	72	34.8	21	0
10537	73	15	9	0
10538	70	15	7	0
10538	72	34.8	1	0
10539	13	6	8	0
10539	21	10	15	0
10539	33	2.5	15	0
10539	49	20	6	0
10540	3	10	60	0
10540	26	31.23	40	0
10540	38	263.5	30	0
10540	68	12.5	35	0
10541	24	4.5	35	0.1
10541	38	263.5	4	0.1
10541	65	21.05	36	0.1
10541	71	21.5	9	0.1
10542	11	21	15	0.05
10542	54	7.45	24	0.05
10543	12	38	30	0.15
10543	23	9	70	0.15
10544	28	45.6	7	0
10544	67	14	7	0
10545	11	21	10	0
10546	7	30	10	0
10546	35	18	30	0
10546	62	49.3	40	0
10547	32	32	24	0.15
10547	36	19	60	0
10548	34	14	10	0.25
10548	41	9.65	14	0
10549	31	12.5	55	0.15
10549	45	9.5	100	0.15
10549	51	53	48	0.15
10550	17	39	8	0.1
10550	19	9.2	10	0
10550	21	10	6	0.1
10550	61	28.5	10	0.1
10551	16	17.45	40	0.15
10551	35	18	20	0.15
10551	44	19.45	40	0
10552	69	36	18	0
10552	75	7.75	30	0
10553	11	21	15	0
10553	16	17.45	14	0
10553	22	21	24	0
10553	31	12.5	30	0
10553	35	18	6	0
10554	16	17.45	30	0.05
10554	23	9	20	0.05
10554	62	49.3	20	0.05
10554	77	13	10	0.05
10555	14	23.25	30	0.2
10555	19	9.2	35	0.2
10555	24	4.5	18	0.2
10555	51	53	20	0.2
10555	56	38	40	0.2
10556	72	34.8	24	0
10557	64	33.25	30	0
10557	75	7.75	20	0
10558	47	9.5	25	0
10558	51	53	20	0
10558	52	7	30	0
10558	53	32.8	18	0
10558	73	15	3	0
10559	41	9.65	12	0.05
10559	55	24	18	0.05
10560	30	25.89	20	0
10560	62	49.3	15	0.25
10561	44	19.45	10	0
10561	51	53	50	0
10562	33	2.5	20	0.1
10562	62	49.3	10	0.1
10563	36	19	25	0
10563	52	7	70	0
10564	17	39	16	0.05
10564	31	12.5	6	0.05
10564	55	24	25	0.05
10565	24	4.5	25	0.1
10565	64	33.25	18	0.1
10566	11	21	35	0.15
10566	18	62.5	18	0.15
10566	76	18	10	0
10567	31	12.5	60	0.2
10567	51	53	3	0
10567	59	55	40	0.2
10568	10	31	5	0
10569	31	12.5	35	0.2
10569	76	18	30	0
10570	11	21	15	0.05
10570	56	38	60	0.05
10571	14	23.25	11	0.15
10571	42	14	28	0.15
10572	16	17.45	12	0.1
10572	32	32	10	0.1
10572	40	18.4	50	0
10572	75	7.75	15	0.1
10573	17	39	18	0
10573	34	14	40	0
10573	53	32.8	25	0
10574	33	2.5	14	0
10574	40	18.4	2	0
10574	62	49.3	10	0
10574	64	33.25	6	0
10575	59	55	12	0
10575	63	43.9	6	0
10575	72	34.8	30	0
10575	76	18	10	0
10576	1	18	10	0
10576	31	12.5	20	0
10576	44	19.45	21	0
10577	39	18	10	0
10577	75	7.75	20	0
10577	77	13	18	0
10578	35	18	20	0
10578	57	19.5	6	0
10579	15	15.5	10	0
10579	75	7.75	21	0
10580	14	23.25	15	0.05
10580	41	9.65	9	0.05
10580	65	21.05	30	0.05
10581	75	7.75	50	0.2
10582	57	19.5	4	0
10582	76	18	14	0
10583	29	123.79	10	0
10583	60	34	24	0.15
10583	69	36	10	0.15
10584	31	12.5	50	0.05
10585	47	9.5	15	0
10586	52	7	4	0.15
10587	26	31.23	6	0
10587	35	18	20	0
10587	77	13	20	0
10588	18	62.5	40	0.2
10588	42	14	100	0.2
10589	35	18	4	0
10590	1	18	20	0
10590	77	13	60	0.05
10591	3	10	14	0
10591	7	30	10	0
10591	54	7.45	50	0
10592	15	15.5	25	0.05
10592	26	31.23	5	0.05
10593	20	81	21	0.2
10593	69	36	20	0.2
10593	76	18	4	0.2
10594	52	7	24	0
10594	58	13.25	30	0
10595	35	18	30	0.25
10595	61	28.5	120	0.25
10595	69	36	65	0.25
10596	56	38	5	0.2
10596	63	43.9	24	0.2
10596	75	7.75	30	0.2
10597	24	4.5	35	0.2
10597	57	19.5	20	0
10597	65	21.05	12	0.2
10598	27	43.9	50	0
10598	71	21.5	9	0
10599	62	49.3	10	0
10600	54	7.45	4	0
10600	73	15	30	0
10601	13	6	60	0
10601	59	55	35	0
10602	77	13	5	0.25
10603	22	21	48	0
10603	49	20	25	0.05
10604	48	12.75	6	0.1
10604	76	18	10	0.1
10605	16	17.45	30	0.05
10605	59	55	20	0.05
10605	60	34	70	0.05
10605	71	21.5	15	0.05
10606	4	22	20	0.2
10606	55	24	20	0.2
10606	62	49.3	10	0.2
10607	7	30	45	0
10607	17	39	100	0
10607	33	2.5	14	0
10607	40	18.4	42	0
10607	72	34.8	12	0
10608	56	38	28	0
10609	1	18	3	0
10609	10	31	10	0
10609	21	10	6	0
10610	36	19	21	0.25
10611	1	18	6	0
10611	2	19	10	0
10611	60	34	15	0
10612	10	31	70	0
10612	36	19	55	0
10612	49	20	18	0
10612	60	34	40	0
10612	76	18	80	0
10613	13	6	8	0.1
10613	75	7.75	40	0
10614	11	21	14	0
10614	21	10	8	0
10614	39	18	5	0
10615	55	24	5	0
10616	38	263.5	15	0.05
10616	56	38	14	0
10616	70	15	15	0.05
10616	71	21.5	15	0.05
10617	59	55	30	0.15
10618	6	25	70	0
10618	56	38	20	0
10618	68	12.5	15	0
10619	21	10	42	0
10619	22	21	40	0
10620	24	4.5	5	0
10620	52	7	5	0
10621	19	9.2	5	0
10621	23	9	10	0
10621	70	15	20	0
10621	71	21.5	15	0
10622	2	19	20	0
10622	68	12.5	18	0.2
10623	14	23.25	21	0
10623	19	9.2	15	0.1
10623	21	10	25	0.1
10623	24	4.5	3	0
10623	35	18	30	0.1
10624	28	45.6	10	0
10624	29	123.79	6	0
10624	44	19.45	10	0
10625	14	23.25	3	0
10625	42	14	5	0
10625	60	34	10	0
10626	53	32.8	12	0
10626	60	34	20	0
10626	71	21.5	20	0
10627	62	49.3	15	0
10627	73	15	35	0.15
10628	1	18	25	0
10629	29	123.79	20	0
10629	64	33.25	9	0
10630	55	24	12	0.05
10630	76	18	35	0
10631	75	7.75	8	0.1
10632	2	19	30	0.05
10632	33	2.5	20	0.05
10633	12	38	36	0.15
10633	13	6	13	0.15
10633	26	31.23	35	0.15
10633	62	49.3	80	0.15
10634	7	30	35	0
10634	18	62.5	50	0
10634	51	53	15	0
10634	75	7.75	2	0
10635	4	22	10	0.1
10635	5	21.35	15	0.1
10635	22	21	40	0
10636	4	22	25	0
10636	58	13.25	6	0
10637	11	21	10	0
10637	50	16.25	25	0.05
10637	56	38	60	0.05
10638	45	9.5	20	0
10638	65	21.05	21	0
10638	72	34.8	60	0
10639	18	62.5	8	0
10640	69	36	20	0.25
10640	70	15	15	0.25
10641	2	19	50	0
10641	40	18.4	60	0
10642	21	10	30	0.2
10642	61	28.5	20	0.2
10643	28	45.6	15	0.25
10643	39	18	21	0.25
10643	46	12	2	0.25
10644	18	62.5	4	0.1
10644	43	46	20	0
10644	46	12	21	0.1
10645	18	62.5	20	0
10645	36	19	15	0
10646	1	18	15	0.25
10646	10	31	18	0.25
10646	71	21.5	30	0.25
10646	77	13	35	0.25
10647	19	9.2	30	0
10647	39	18	20	0
10648	22	21	15	0
10648	24	4.5	15	0.15
10649	28	45.6	20	0
10649	72	34.8	15	0
10650	30	25.89	30	0
10650	53	32.8	25	0.05
10650	54	7.45	30	0
10651	19	9.2	12	0.25
10651	22	21	20	0.25
10652	30	25.89	2	0.25
10652	42	14	20	0
10653	16	17.45	30	0.1
10653	60	34	20	0.1
10654	4	22	12	0.1
10654	39	18	20	0.1
10654	54	7.45	6	0.1
10655	41	9.65	20	0.2
10656	14	23.25	3	0.1
10656	44	19.45	28	0.1
10656	47	9.5	6	0.1
10657	15	15.5	50	0
10657	41	9.65	24	0
10657	46	12	45	0
10657	47	9.5	10	0
10657	56	38	45	0
10657	60	34	30	0
10658	21	10	60	0
10658	40	18.4	70	0.05
10658	60	34	55	0.05
10658	77	13	70	0.05
10659	31	12.5	20	0.05
10659	40	18.4	24	0.05
10659	70	15	40	0.05
10660	20	81	21	0
10661	39	18	3	0.2
10661	58	13.25	49	0.2
10662	68	12.5	10	0
10663	40	18.4	30	0.05
10663	42	14	30	0.05
10663	51	53	20	0.05
10664	10	31	24	0.15
10664	56	38	12	0.15
10664	65	21.05	15	0.15
10665	51	53	20	0
10665	59	55	1	0
10665	76	18	10	0
10666	29	123.79	36	0
10666	65	21.05	10	0
10667	69	36	45	0.2
10667	71	21.5	14	0.2
10668	31	12.5	8	0.1
10668	55	24	4	0.1
10668	64	33.25	15	0.1
10669	36	19	30	0
10670	23	9	32	0
10670	46	12	60	0
10670	67	14	25	0
10670	73	15	50	0
10670	75	7.75	25	0
10671	16	17.45	10	0
10671	62	49.3	10	0
10671	65	21.05	12	0
10672	38	263.5	15	0.1
10672	71	21.5	12	0
10673	16	17.45	3	0
10673	42	14	6	0
10673	43	46	6	0
10674	23	9	5	0
10675	14	23.25	30	0
10675	53	32.8	10	0
10675	58	13.25	30	0
10676	10	31	2	0
10676	19	9.2	7	0
10676	44	19.45	21	0
10677	26	31.23	30	0.15
10677	33	2.5	8	0.15
10678	12	38	100	0
10678	33	2.5	30	0
10678	41	9.65	120	0
10678	54	7.45	30	0
10679	59	55	12	0
10680	16	17.45	50	0.25
10680	31	12.5	20	0.25
10680	42	14	40	0.25
10681	19	9.2	30	0.1
10681	21	10	12	0.1
10681	64	33.25	28	0
10682	33	2.5	30	0
10682	66	17	4	0
10682	75	7.75	30	0
10683	52	7	9	0
10684	40	18.4	20	0
10684	47	9.5	40	0
10684	60	34	30	0
10685	10	31	20	0
10685	41	9.65	4	0
10685	47	9.5	15	0
10686	17	39	30	0.2
10686	26	31.23	15	0
10687	9	97	50	0.25
10687	29	123.79	10	0
10687	36	19	6	0.25
10688	10	31	18	0.1
10688	28	45.6	60	0.1
10688	34	14	14	0
10689	1	18	35	0.25
10690	56	38	20	0.25
10690	77	13	30	0.25
10691	1	18	30	0
10691	29	123.79	40	0
10691	43	46	40	0
10691	44	19.45	24	0
10691	62	49.3	48	0
10692	63	43.9	20	0
10693	9	97	6	0
10693	54	7.45	60	0.15
10693	69	36	30	0.15
10693	73	15	15	0.15
10694	7	30	90	0
10694	59	55	25	0
10694	70	15	50	0
10695	8	40	10	0
10695	12	38	4	0
10695	24	4.5	20	0
10696	17	39	20	0
10696	46	12	18	0
10697	19	9.2	7	0.25
10697	35	18	9	0.25
10697	58	13.25	30	0.25
10697	70	15	30	0.25
10698	11	21	15	0
10698	17	39	8	0.05
10698	29	123.79	12	0.05
10698	65	21.05	65	0.05
10698	70	15	8	0.05
10699	47	9.5	12	0
10700	1	18	5	0.2
10700	34	14	12	0.2
10700	68	12.5	40	0.2
10700	71	21.5	60	0.2
10701	59	55	42	0.15
10701	71	21.5	20	0.15
10701	76	18	35	0.15
10702	3	10	6	0
10702	76	18	15	0
10703	2	19	5	0
10703	59	55	35	0
10703	73	15	35	0
10704	4	22	6	0
10704	24	4.5	35	0
10704	48	12.75	24	0
10705	31	12.5	20	0
10705	32	32	4	0
10706	16	17.45	20	0
10706	43	46	24	0
10706	59	55	8	0
10707	55	24	21	0
10707	57	19.5	40	0
10707	70	15	28	0.15
10708	5	21.35	4	0
10708	36	19	5	0
10709	8	40	40	0
10709	51	53	28	0
10709	60	34	10	0
10710	19	9.2	5	0
10710	47	9.5	5	0
10711	19	9.2	12	0
10711	41	9.65	42	0
10711	53	32.8	120	0
10712	53	32.8	3	0.05
10712	56	38	30	0
10713	10	31	18	0
10713	26	31.23	30	0
10713	45	9.5	110	0
10713	46	12	24	0
10714	2	19	30	0.25
10714	17	39	27	0.25
10714	47	9.5	50	0.25
10714	56	38	18	0.25
10714	58	13.25	12	0.25
10715	10	31	21	0
10715	71	21.5	30	0
10716	21	10	5	0
10716	51	53	7	0
10716	61	28.5	10	0
10717	21	10	32	0.05
10717	54	7.45	15	0
10717	69	36	25	0.05
10718	12	38	36	0
10718	16	17.45	20	0
10718	36	19	40	0
10718	62	49.3	20	0
10719	18	62.5	12	0.25
10719	30	25.89	3	0.25
10719	54	7.45	40	0.25
10720	35	18	21	0
10720	71	21.5	8	0
10721	44	19.45	50	0.05
10722	2	19	3	0
10722	31	12.5	50	0
10722	68	12.5	45	0
10722	75	7.75	42	0
10723	26	31.23	15	0
10724	10	31	16	0
10724	61	28.5	5	0
10725	41	9.65	12	0
10725	52	7	4	0
10725	55	24	6	0
10726	4	22	25	0
10726	11	21	5	0
10727	17	39	20	0.05
10727	56	38	10	0.05
10727	59	55	10	0.05
10728	30	25.89	15	0
10728	40	18.4	6	0
10728	55	24	12	0
10728	60	34	15	0
10729	1	18	50	0
10729	21	10	30	0
10729	50	16.25	40	0
10730	16	17.45	15	0.05
10730	31	12.5	3	0.05
10730	65	21.05	10	0.05
10731	21	10	40	0.05
10731	51	53	30	0.05
10732	76	18	20	0
10733	14	23.25	16	0
10733	28	45.6	20	0
10733	52	7	25	0
10734	6	25	30	0
10734	30	25.89	15	0
10734	76	18	20	0
10735	61	28.5	20	0.1
10735	77	13	2	0.1
10736	65	21.05	40	0
10736	75	7.75	20	0
10737	13	6	4	0
10737	41	9.65	12	0
10738	16	17.45	3	0
10739	36	19	6	0
10739	52	7	18	0
10740	28	45.6	5	0.2
10740	35	18	35	0.2
10740	45	9.5	40	0.2
10740	56	38	14	0.2
10741	2	19	15	0.2
10742	3	10	20	0
10742	60	34	50	0
10742	72	34.8	35	0
10743	46	12	28	0.05
10744	40	18.4	50	0.2
10745	18	62.5	24	0
10745	44	19.45	16	0
10745	59	55	45	0
10745	72	34.8	7	0
10746	13	6	6	0
10746	42	14	28	0
10746	62	49.3	9	0
10746	69	36	40	0
10747	31	12.5	8	0
10747	41	9.65	35	0
10747	63	43.9	9	0
10747	69	36	30	0
10748	23	9	44	0
10748	40	18.4	40	0
10748	56	38	28	0
10749	56	38	15	0
10749	59	55	6	0
10749	76	18	10	0
10750	14	23.25	5	0.15
10750	45	9.5	40	0.15
10750	59	55	25	0.15
10751	26	31.23	12	0.1
10751	30	25.89	30	0
10751	50	16.25	20	0.1
10751	73	15	15	0
10752	1	18	8	0
10752	69	36	3	0
10753	45	9.5	4	0
10753	74	10	5	0
10754	40	18.4	3	0
10755	47	9.5	30	0.25
10755	56	38	30	0.25
10755	57	19.5	14	0.25
10755	69	36	25	0.25
10756	18	62.5	21	0.2
10756	36	19	20	0.2
10756	68	12.5	6	0.2
10756	69	36	20	0.2
10757	34	14	30	0
10757	59	55	7	0
10757	62	49.3	30	0
10757	64	33.25	24	0
10758	26	31.23	20	0
10758	52	7	60	0
10758	70	15	40	0
10759	32	32	10	0
10760	25	14	12	0.25
10760	27	43.9	40	0
10760	43	46	30	0.25
10761	25	14	35	0.25
10761	75	7.75	18	0
10762	39	18	16	0
10762	47	9.5	30	0
10762	51	53	28	0
10762	56	38	60	0
10763	21	10	40	0
10763	22	21	6	0
10763	24	4.5	20	0
10764	3	10	20	0.1
10764	39	18	130	0.1
10765	65	21.05	80	0.1
10766	2	19	40	0
10766	7	30	35	0
10766	68	12.5	40	0
10767	42	14	2	0
10768	22	21	4	0
10768	31	12.5	50	0
10768	60	34	15	0
10768	71	21.5	12	0
10769	41	9.65	30	0.05
10769	52	7	15	0.05
10769	61	28.5	20	0
10769	62	49.3	15	0
10770	11	21	15	0.25
10771	71	21.5	16	0
10772	29	123.79	18	0
10772	59	55	25	0
10773	17	39	33	0
10773	31	12.5	70	0.2
10773	75	7.75	7	0.2
10774	31	12.5	2	0.25
10774	66	17	50	0
10775	10	31	6	0
10775	67	14	3	0
10776	31	12.5	16	0.05
10776	42	14	12	0.05
10776	45	9.5	27	0.05
10776	51	53	120	0.05
10777	42	14	20	0.2
10778	41	9.65	10	0
10779	16	17.45	20	0
10779	62	49.3	20	0
10780	70	15	35	0
10780	77	13	15	0
10781	54	7.45	3	0.2
10781	56	38	20	0.2
10781	74	10	35	0
10782	31	12.5	1	0
10783	31	12.5	10	0
10783	38	263.5	5	0
10784	36	19	30	0
10784	39	18	2	0.15
10784	72	34.8	30	0.15
10785	10	31	10	0
10785	75	7.75	10	0
10786	8	40	30	0.2
10786	30	25.89	15	0.2
10786	75	7.75	42	0.2
10787	2	19	15	0.05
10787	29	123.79	20	0.05
10788	19	9.2	50	0.05
10788	75	7.75	40	0.05
10789	18	62.5	30	0
10789	35	18	15	0
10789	63	43.9	30	0
10789	68	12.5	18	0
10790	7	30	3	0.15
10790	56	38	20	0.15
10791	29	123.79	14	0.05
10791	41	9.65	20	0.05
10792	2	19	10	0
10792	54	7.45	3	0
10792	68	12.5	15	0
10793	41	9.65	14	0
10793	52	7	8	0
10794	14	23.25	15	0.2
10794	54	7.45	6	0.2
10795	16	17.45	65	0
10795	17	39	35	0.25
10796	26	31.23	21	0.2
10796	44	19.45	10	0
10796	64	33.25	35	0.2
10796	69	36	24	0.2
10797	11	21	20	0
10798	62	49.3	2	0
10798	72	34.8	10	0
10799	13	6	20	0.15
10799	24	4.5	20	0.15
10799	59	55	25	0
10800	11	21	50	0.1
10800	51	53	10	0.1
10800	54	7.45	7	0.1
10801	17	39	40	0.25
10801	29	123.79	20	0.25
10802	30	25.89	25	0.25
10802	51	53	30	0.25
10802	55	24	60	0.25
10802	62	49.3	5	0.25
10803	19	9.2	24	0.05
10803	25	14	15	0.05
10803	59	55	15	0.05
10804	10	31	36	0
10804	28	45.6	24	0
10804	49	20	4	0.15
10805	34	14	10	0
10805	38	263.5	10	0
10806	2	19	20	0.25
10806	65	21.05	2	0
10806	74	10	15	0.25
10807	40	18.4	1	0
10808	56	38	20	0.15
10808	76	18	50	0.15
10809	52	7	20	0
10810	13	6	7	0
10810	25	14	5	0
10810	70	15	5	0
10811	19	9.2	15	0
10811	23	9	18	0
10811	40	18.4	30	0
10812	31	12.5	16	0.1
10812	72	34.8	40	0.1
10812	77	13	20	0
10813	2	19	12	0.2
10813	46	12	35	0
10814	41	9.65	20	0
10814	43	46	20	0.15
10814	48	12.75	8	0.15
10814	61	28.5	30	0.15
10815	33	2.5	16	0
10816	38	263.5	30	0.05
10816	62	49.3	20	0.05
10817	26	31.23	40	0.15
10817	38	263.5	30	0
10817	40	18.4	60	0.15
10817	62	49.3	25	0.15
10818	32	32	20	0
10818	41	9.65	20	0
10819	43	46	7	0
10819	75	7.75	20	0
10820	56	38	30	0
10821	35	18	20	0
10821	51	53	6	0
10822	62	49.3	3	0
10822	70	15	6	0
10823	11	21	20	0.1
10823	57	19.5	15	0
10823	59	55	40	0.1
10823	77	13	15	0.1
10824	41	9.65	12	0
10824	70	15	9	0
10825	26	31.23	12	0
10825	53	32.8	20	0
10826	31	12.5	35	0
10826	57	19.5	15	0
10827	10	31	15	0
10827	39	18	21	0
10828	20	81	5	0
10828	38	263.5	2	0
10829	2	19	10	0
10829	8	40	20	0
10829	13	6	10	0
10829	60	34	21	0
10830	6	25	6	0
10830	39	18	28	0
10830	60	34	30	0
10830	68	12.5	24	0
10831	19	9.2	2	0
10831	35	18	8	0
10831	38	263.5	8	0
10831	43	46	9	0
10832	13	6	3	0.2
10832	25	14	10	0.2
10832	44	19.45	16	0.2
10832	64	33.25	3	0
10833	7	30	20	0.1
10833	31	12.5	9	0.1
10833	53	32.8	9	0.1
10834	29	123.79	8	0.05
10834	30	25.89	20	0.05
10835	59	55	15	0
10835	77	13	2	0.2
10836	22	21	52	0
10836	35	18	6	0
10836	57	19.5	24	0
10836	60	34	60	0
10836	64	33.25	30	0
10837	13	6	6	0
10837	40	18.4	25	0
10837	47	9.5	40	0.25
10837	76	18	21	0.25
10838	1	18	4	0.25
10838	18	62.5	25	0.25
10838	36	19	50	0.25
10839	58	13.25	30	0.1
10839	72	34.8	15	0.1
10840	25	14	6	0.2
10840	39	18	10	0.2
10841	10	31	16	0
10841	56	38	30	0
10841	59	55	50	0
10841	77	13	15	0
10842	11	21	15	0
10842	43	46	5	0
10842	68	12.5	20	0
10842	70	15	12	0
10843	51	53	4	0.25
10844	22	21	35	0
10845	23	9	70	0.1
10845	35	18	25	0.1
10845	42	14	42	0.1
10845	58	13.25	60	0.1
10845	64	33.25	48	0
10846	4	22	21	0
10846	70	15	30	0
10846	74	10	20	0
10847	1	18	80	0.2
10847	19	9.2	12	0.2
10847	37	26	60	0.2
10847	45	9.5	36	0.2
10847	60	34	45	0.2
10847	71	21.5	55	0.2
10848	5	21.35	30	0
10848	9	97	3	0
10849	3	10	49	0
10849	26	31.23	18	0.15
10850	25	14	20	0.15
10850	33	2.5	4	0.15
10850	70	15	30	0.15
10851	2	19	5	0.05
10851	25	14	10	0.05
10851	57	19.5	10	0.05
10851	59	55	42	0.05
10852	2	19	15	0
10852	17	39	6	0
10852	62	49.3	50	0
10853	18	62.5	10	0
10854	10	31	100	0.15
10854	13	6	65	0.15
10855	16	17.45	50	0
10855	31	12.5	14	0
10855	56	38	24	0
10855	65	21.05	15	0.15
10856	2	19	20	0
10856	42	14	20	0
10857	3	10	30	0
10857	26	31.23	35	0.25
10857	29	123.79	10	0.25
10858	7	30	5	0
10858	27	43.9	10	0
10858	70	15	4	0
10859	24	4.5	40	0.25
10859	54	7.45	35	0.25
10859	64	33.25	30	0.25
10860	51	53	3	0
10860	76	18	20	0
10861	17	39	42	0
10861	18	62.5	20	0
10861	21	10	40	0
10861	33	2.5	35	0
10861	62	49.3	3	0
10862	11	21	25	0
10862	52	7	8	0
10863	1	18	20	0.15
10863	58	13.25	12	0.15
10864	35	18	4	0
10864	67	14	15	0
10865	38	263.5	60	0.05
10865	39	18	80	0.05
10866	2	19	21	0.25
10866	24	4.5	6	0.25
10866	30	25.89	40	0.25
10867	53	32.8	3	0
10868	26	31.23	20	0
10868	35	18	30	0
10868	49	20	42	0.1
10869	1	18	40	0
10869	11	21	10	0
10869	23	9	50	0
10869	68	12.5	20	0
10870	35	18	3	0
10870	51	53	2	0
10871	6	25	50	0.05
10871	16	17.45	12	0.05
10871	17	39	16	0.05
10872	55	24	10	0.05
10872	62	49.3	20	0.05
10872	64	33.25	15	0.05
10872	65	21.05	21	0.05
10873	21	10	20	0
10873	28	45.6	3	0
10874	10	31	10	0
10875	19	9.2	25	0
10875	47	9.5	21	0.1
10875	49	20	15	0
10876	46	12	21	0
10876	64	33.25	20	0
10877	16	17.45	30	0.25
10877	18	62.5	25	0
10878	20	81	20	0.05
10879	40	18.4	12	0
10879	65	21.05	10	0
10879	76	18	10	0
10880	23	9	30	0.2
10880	61	28.5	30	0.2
10880	70	15	50	0.2
10881	73	15	10	0
10882	42	14	25	0
10882	49	20	20	0.15
10882	54	7.45	32	0.15
10883	24	4.5	8	0
10884	21	10	40	0.05
10884	56	38	21	0.05
10884	65	21.05	12	0.05
10885	2	19	20	0
10885	24	4.5	12	0
10885	70	15	30	0
10885	77	13	25	0
10886	10	31	70	0
10886	31	12.5	35	0
10886	77	13	40	0
10887	25	14	5	0
10888	2	19	20	0
10888	68	12.5	18	0
10889	11	21	40	0
10889	38	263.5	40	0
10890	17	39	15	0
10890	34	14	10	0
10890	41	9.65	14	0
10891	30	25.89	15	0.05
10892	59	55	40	0.05
10893	8	40	30	0
10893	24	4.5	10	0
10893	29	123.79	24	0
10893	30	25.89	35	0
10893	36	19	20	0
10894	13	6	28	0.05
10894	69	36	50	0.05
10894	75	7.75	120	0.05
10895	24	4.5	110	0
10895	39	18	45	0
10895	40	18.4	91	0
10895	60	34	100	0
10896	45	9.5	15	0
10896	56	38	16	0
10897	29	123.79	80	0
10897	30	25.89	36	0
10898	13	6	5	0
10899	39	18	8	0.15
10900	70	15	3	0.25
10901	41	9.65	30	0
10901	71	21.5	30	0
10902	55	24	30	0.15
10902	62	49.3	6	0.15
10903	13	6	40	0
10903	65	21.05	21	0
10903	68	12.5	20	0
10904	58	13.25	15	0
10904	62	49.3	35	0
10905	1	18	20	0.05
10906	61	28.5	15	0
10907	75	7.75	14	0
10908	7	30	20	0.05
10908	52	7	14	0.05
10909	7	30	12	0
10909	16	17.45	15	0
10909	41	9.65	5	0
10910	19	9.2	12	0
10910	49	20	10	0
10910	61	28.5	5	0
10911	1	18	10	0
10911	17	39	12	0
10911	67	14	15	0
10912	11	21	40	0.25
10912	29	123.79	60	0.25
10913	4	22	30	0.25
10913	33	2.5	40	0.25
10913	58	13.25	15	0
10914	71	21.5	25	0
10915	17	39	10	0
10915	33	2.5	30	0
10915	54	7.45	10	0
10916	16	17.45	6	0
10916	32	32	6	0
10916	57	19.5	20	0
10917	30	25.89	1	0
10917	60	34	10	0
10918	1	18	60	0.25
10918	60	34	25	0.25
10919	16	17.45	24	0
10919	25	14	24	0
10919	40	18.4	20	0
10920	50	16.25	24	0
10921	35	18	10	0
10921	63	43.9	40	0
10922	17	39	15	0
10922	24	4.5	35	0
10923	42	14	10	0.2
10923	43	46	10	0.2
10923	67	14	24	0.2
10924	10	31	20	0.1
10924	28	45.6	30	0.1
10924	75	7.75	6	0
10925	36	19	25	0.15
10925	52	7	12	0.15
10926	11	21	2	0
10926	13	6	10	0
10926	19	9.2	7	0
10926	72	34.8	10	0
10927	20	81	5	0
10927	52	7	5	0
10927	76	18	20	0
10928	47	9.5	5	0
10928	76	18	5	0
10929	21	10	60	0
10929	75	7.75	49	0
10929	77	13	15	0
10930	21	10	36	0
10930	27	43.9	25	0
10930	55	24	25	0.2
10930	58	13.25	30	0.2
10931	13	6	42	0.15
10931	57	19.5	30	0
10932	16	17.45	30	0.1
10932	62	49.3	14	0.1
10932	72	34.8	16	0
10932	75	7.75	20	0.1
10933	53	32.8	2	0
10933	61	28.5	30	0
10934	6	25	20	0
10935	1	18	21	0
10935	18	62.5	4	0.25
10935	23	9	8	0.25
10936	36	19	30	0.2
10937	28	45.6	8	0
10937	34	14	20	0
10938	13	6	20	0.25
10938	43	46	24	0.25
10938	60	34	49	0.25
10938	71	21.5	35	0.25
10939	2	19	10	0.15
10939	67	14	40	0.15
10940	7	30	8	0
10940	13	6	20	0
10941	31	12.5	44	0.25
10941	62	49.3	30	0.25
10941	68	12.5	80	0.25
10941	72	34.8	50	0
10942	49	20	28	0
10943	13	6	15	0
10943	22	21	21	0
10943	46	12	15	0
10944	11	21	5	0.25
10944	44	19.45	18	0.25
10944	56	38	18	0
10945	13	6	20	0
10945	31	12.5	10	0
10946	10	31	25	0
10946	24	4.5	25	0
10946	77	13	40	0
10947	59	55	4	0
10948	50	16.25	9	0
10948	51	53	40	0
10948	55	24	4	0
10949	6	25	12	0
10949	10	31	30	0
10949	17	39	6	0
10949	62	49.3	60	0
10950	4	22	5	0
10951	33	2.5	15	0.05
10951	41	9.65	6	0.05
10951	75	7.75	50	0.05
10952	6	25	16	0.05
10952	28	45.6	2	0
10953	20	81	50	0.05
10953	31	12.5	50	0.05
10954	16	17.45	28	0.15
10954	31	12.5	25	0.15
10954	45	9.5	30	0
10954	60	34	24	0.15
10955	75	7.75	12	0.2
10956	21	10	12	0
10956	47	9.5	14	0
10956	51	53	8	0
10957	30	25.89	30	0
10957	35	18	40	0
10957	64	33.25	8	0
10958	5	21.35	20	0
10958	7	30	6	0
10958	72	34.8	5	0
10959	75	7.75	20	0.15
10960	24	4.5	10	0.25
10960	41	9.65	24	0
10961	52	7	6	0.05
10961	76	18	60	0
10962	7	30	45	0
10962	13	6	77	0
10962	53	32.8	20	0
10962	69	36	9	0
10962	76	18	44	0
10963	60	34	2	0.15
10964	18	62.5	6	0
10964	38	263.5	5	0
10964	69	36	10	0
10965	51	53	16	0
10966	37	26	8	0
10966	56	38	12	0.15
10966	62	49.3	12	0.15
10967	19	9.2	12	0
10967	49	20	40	0
10968	12	38	30	0
10968	24	4.5	30	0
10968	64	33.25	4	0
10969	46	12	9	0
10970	52	7	40	0.2
10971	29	123.79	14	0
10972	17	39	6	0
10972	33	2.5	7	0
10973	26	31.23	5	0
10973	41	9.65	6	0
10973	75	7.75	10	0
10974	63	43.9	10	0
10975	8	40	16	0
10975	75	7.75	10	0
10976	28	45.6	20	0
10977	39	18	30	0
10977	47	9.5	30	0
10977	51	53	10	0
10977	63	43.9	20	0
10978	8	40	20	0.15
10978	21	10	40	0.15
10978	40	18.4	10	0
10978	44	19.45	6	0.15
10979	7	30	18	0
10979	12	38	20	0
10979	24	4.5	80	0
10979	27	43.9	30	0
10979	31	12.5	24	0
10979	63	43.9	35	0
10980	75	7.75	40	0.2
10981	38	263.5	60	0
10982	7	30	20	0
10982	43	46	9	0
10983	13	6	84	0.15
10983	57	19.5	15	0
10984	16	17.45	55	0
10984	24	4.5	20	0
10984	36	19	40	0
10985	16	17.45	36	0.1
10985	18	62.5	8	0.1
10985	32	32	35	0.1
10986	11	21	30	0
10986	20	81	15	0
10986	76	18	10	0
10986	77	13	15	0
10987	7	30	60	0
10987	43	46	6	0
10987	72	34.8	20	0
10988	7	30	60	0
10988	62	49.3	40	0.1
10989	6	25	40	0
10989	11	21	15	0
10989	41	9.65	4	0
10990	21	10	65	0
10990	34	14	60	0.15
10990	55	24	65	0.15
10990	61	28.5	66	0.15
10991	2	19	50	0.2
10991	70	15	20	0.2
10991	76	18	90	0.2
10992	72	34.8	2	0
10993	29	123.79	50	0.25
10993	41	9.65	35	0.25
10994	59	55	18	0.05
10995	51	53	20	0
10995	60	34	4	0
10996	42	14	40	0
10997	32	32	50	0
10997	46	12	20	0.25
10997	52	7	20	0.25
10998	24	4.5	12	0
10998	61	28.5	7	0
10998	74	10	20	0
10998	75	7.75	30	0
10999	41	9.65	20	0.05
10999	51	53	15	0.05
10999	77	13	21	0.05
11000	4	22	25	0.25
11000	24	4.5	30	0.25
11000	77	13	30	0
11001	7	30	60	0
11001	22	21	25	0
11001	46	12	25	0
11001	55	24	6	0
11002	13	6	56	0
11002	35	18	15	0.15
11002	42	14	24	0.15
11002	55	24	40	0
11003	1	18	4	0
11003	40	18.4	10	0
11003	52	7	10	0
11004	26	31.23	6	0
11004	76	18	6	0
11005	1	18	2	0
11005	59	55	10	0
11006	1	18	8	0
11006	29	123.79	2	0.25
11007	8	40	30	0
11007	29	123.79	10	0
11007	42	14	14	0
11008	28	45.6	70	0.05
11008	34	14	90	0.05
11008	71	21.5	21	0
11009	24	4.5	12	0
11009	36	19	18	0.25
11009	60	34	9	0
11010	7	30	20	0
11010	24	4.5	10	0
11011	58	13.25	40	0.05
11011	71	21.5	20	0
11012	19	9.2	50	0.05
11012	60	34	36	0.05
11012	71	21.5	60	0.05
11013	23	9	10	0
11013	42	14	4	0
11013	45	9.5	20	0
11013	68	12.5	2	0
11014	41	9.65	28	0.1
11015	30	25.89	15	0
11015	77	13	18	0
11016	31	12.5	15	0
11016	36	19	16	0
11017	3	10	25	0
11017	59	55	110	0
11017	70	15	30	0
11018	12	38	20	0
11018	18	62.5	10	0
11018	56	38	5	0
11019	46	12	3	0
11019	49	20	2	0
11020	10	31	24	0.15
11021	2	19	11	0.25
11021	20	81	15	0
11021	26	31.23	63	0
11021	51	53	44	0.25
11021	72	34.8	35	0
11022	19	9.2	35	0
11022	69	36	30	0
11023	7	30	4	0
11023	43	46	30	0
11024	26	31.23	12	0
11024	33	2.5	30	0
11024	65	21.05	21	0
11024	71	21.5	50	0
11025	1	18	10	0.1
11025	13	6	20	0.1
11026	18	62.5	8	0
11026	51	53	10	0
11027	24	4.5	30	0.25
11027	62	49.3	21	0.25
11028	55	24	35	0
11028	59	55	24	0
11029	56	38	20	0
11029	63	43.9	12	0
11030	2	19	100	0.25
11030	5	21.35	70	0
11030	29	123.79	60	0.25
11030	59	55	100	0.25
11031	1	18	45	0
11031	13	6	80	0
11031	24	4.5	21	0
11031	64	33.25	20	0
11031	71	21.5	16	0
11032	36	19	35	0
11032	38	263.5	25	0
11032	59	55	30	0
11033	53	32.8	70	0.1
11033	69	36	36	0.1
11034	21	10	15	0.1
11034	44	19.45	12	0
11034	61	28.5	6	0
11035	1	18	10	0
11035	35	18	60	0
11035	42	14	30	0
11035	54	7.45	10	0
11036	13	6	7	0
11036	59	55	30	0
11037	70	15	4	0
11038	40	18.4	5	0.2
11038	52	7	2	0
11038	71	21.5	30	0
11039	28	45.6	20	0
11039	35	18	24	0
11039	49	20	60	0
11039	57	19.5	28	0
11040	21	10	20	0
11041	2	19	30	0.2
11041	63	43.9	30	0
11042	44	19.45	15	0
11042	61	28.5	4	0
11043	11	21	10	0
11044	62	49.3	12	0
11045	33	2.5	15	0
11045	51	53	24	0
11046	12	38	20	0.05
11046	32	32	15	0.05
11046	35	18	18	0.05
11047	1	18	25	0.25
11047	5	21.35	30	0.25
11048	68	12.5	42	0
11049	2	19	10	0.2
11049	12	38	4	0.2
11050	76	18	50	0.1
11051	24	4.5	10	0.2
11052	43	46	30	0.2
11052	61	28.5	10	0.2
11053	18	62.5	35	0.2
11053	32	32	20	0
11053	64	33.25	25	0.2
11054	33	2.5	10	0
11054	67	14	20	0
11055	24	4.5	15	0
11055	25	14	15	0
11055	51	53	20	0
11055	57	19.5	20	0
11056	7	30	40	0
11056	55	24	35	0
11056	60	34	50	0
11057	70	15	3	0
11058	21	10	3	0
11058	60	34	21	0
11058	61	28.5	4	0
11059	13	6	30	0
11059	17	39	12	0
11059	60	34	35	0
11060	60	34	4	0
11060	77	13	10	0
11061	60	34	15	0
11062	53	32.8	10	0.2
11062	70	15	12	0.2
11063	34	14	30	0
11063	40	18.4	40	0.1
11063	41	9.65	30	0.1
11064	17	39	77	0.1
11064	41	9.65	12	0
11064	53	32.8	25	0.1
11064	55	24	4	0.1
11064	68	12.5	55	0
11065	30	25.89	4	0.25
11065	54	7.45	20	0.25
11066	16	17.45	3	0
11066	19	9.2	42	0
11066	34	14	35	0
11067	41	9.65	9	0
11068	28	45.6	8	0.15
11068	43	46	36	0.15
11068	77	13	28	0.15
11069	39	18	20	0
11070	1	18	40	0.15
11070	2	19	20	0.15
11070	16	17.45	30	0.15
11070	31	12.5	20	0
11071	7	30	15	0.05
11071	13	6	10	0.05
11072	2	19	8	0
11072	41	9.65	40	0
11072	50	16.25	22	0
11072	64	33.25	130	0
11073	11	21	10	0
11073	24	4.5	20	0
11074	16	17.45	14	0.05
11075	2	19	10	0.15
11075	46	12	30	0.15
11075	76	18	2	0.15
11076	6	25	20	0.25
11076	14	23.25	20	0.25
11076	19	9.2	10	0.25
11077	2	19	24	0.2
11077	3	10	4	0
11077	4	22	1	0
11077	6	25	1	0.02
11077	7	30	1	0.05
11077	8	40	2	0.1
11077	10	31	1	0
11077	12	38	2	0.05
11077	13	6	4	0
11077	14	23.25	1	0.03
11077	16	17.45	2	0.03
11077	20	81	1	0.04
11077	23	9	2	0
11077	32	32	1	0
11077	39	18	2	0.05
11077	41	9.65	3	0
11077	46	12	3	0.02
11077	52	7	2	0
11077	55	24	2	0
11077	60	34	2	0.06
11077	64	33.25	2	0.03
11077	66	17	1	0
11077	73	15	2	0.01
11077	75	7.75	4	0
11077	77	13	2	0
\.


--
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.orders (order_id, customer_id, employee_id, order_date, required_date, shipped_date, ship_via, freight, ship_name, ship_address, ship_city, ship_region, ship_postal_code, ship_country) FROM stdin;
10248	VINET	5	1996-07-04	1996-08-01	1996-07-16	3	32.38	Vins et alcools Chevalier	59 rue de l'Abbaye	Reims	\N	51100	France
10249	TOMSP	6	1996-07-05	1996-08-16	1996-07-10	1	11.61	Toms Spezialitäten	Luisenstr. 48	Münster	\N	44087	Germany
10251	VICTE	3	1996-07-08	1996-08-05	1996-07-15	1	41.34	Victuailles en stock	2, rue du Commerce	Lyon	\N	69004	France
10252	SUPRD	4	1996-07-09	1996-08-06	1996-07-11	2	51.3	Suprêmes délices	Boulevard Tirou, 255	Charleroi	\N	B-6000	Belgium
10253	HANAR	3	1996-07-10	1996-07-24	1996-07-16	2	58.17	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
10254	CHOPS	5	1996-07-11	1996-08-08	1996-07-23	2	22.98	Chop-suey Chinese	Hauptstr. 31	Bern	\N	3012	Switzerland
10255	RICSU	9	1996-07-12	1996-08-09	1996-07-15	3	148.33	Richter Supermarkt	Starenweg 5	Genève	\N	1204	Switzerland
10256	WELLI	3	1996-07-15	1996-08-12	1996-07-17	2	13.97	Wellington Importadora	Rua do Mercado, 12	Resende	SP	08737-363	Brazil
10257	HILAA	4	1996-07-16	1996-08-13	1996-07-22	3	81.91	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10258	ERNSH	1	1996-07-17	1996-08-14	1996-07-23	1	140.51	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10259	CENTC	4	1996-07-18	1996-08-15	1996-07-25	3	3.25	Centro comercial Moctezuma	Sierras de Granada 9993	México D.F.	\N	05022	Mexico
10260	OTTIK	4	1996-07-19	1996-08-16	1996-07-29	1	55.09	Ottilies Käseladen	Mehrheimerstr. 369	Köln	\N	50739	Germany
10261	QUEDE	4	1996-07-19	1996-08-16	1996-07-30	2	3.05	Que Delícia	Rua da Panificadora, 12	Rio de Janeiro	RJ	02389-673	Brazil
10262	RATTC	8	1996-07-22	1996-08-19	1996-07-25	3	48.29	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10263	ERNSH	9	1996-07-23	1996-08-20	1996-07-31	3	146.06	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10264	FOLKO	6	1996-07-24	1996-08-21	1996-08-23	3	3.67	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10265	BLONP	2	1996-07-25	1996-08-22	1996-08-12	1	55.28	Blondel père et fils	24, place Kléber	Strasbourg	\N	67000	France
10266	WARTH	3	1996-07-26	1996-09-06	1996-07-31	3	25.73	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10267	FRANK	4	1996-07-29	1996-08-26	1996-08-06	1	208.58	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10268	GROSR	8	1996-07-30	1996-08-27	1996-08-02	3	66.29	GROSELLA-Restaurante	5ª Ave. Los Palos Grandes	Caracas	DF	1081	Venezuela
10269	WHITC	5	1996-07-31	1996-08-14	1996-08-09	1	4.56	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
10270	WARTH	1	1996-08-01	1996-08-29	1996-08-02	1	136.54	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10271	SPLIR	6	1996-08-01	1996-08-29	1996-08-30	2	4.54	Split Rail Beer & Ale	P.O. Box 555	Lander	WY	82520	USA
10272	RATTC	6	1996-08-02	1996-08-30	1996-08-06	2	98.03	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10273	QUICK	3	1996-08-05	1996-09-02	1996-08-12	3	76.07	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10274	VINET	6	1996-08-06	1996-09-03	1996-08-16	1	6.01	Vins et alcools Chevalier	59 rue de l'Abbaye	Reims	\N	51100	France
10275	MAGAA	1	1996-08-07	1996-09-04	1996-08-09	1	26.93	Magazzini Alimentari Riuniti	Via Ludovico il Moro 22	Bergamo	\N	24100	Italy
10276	TORTU	8	1996-08-08	1996-08-22	1996-08-14	3	13.84	Tortuga Restaurante	Avda. Azteca 123	México D.F.	\N	05033	Mexico
10277	MORGK	2	1996-08-09	1996-09-06	1996-08-13	3	125.77	Morgenstern Gesundkost	Heerstr. 22	Leipzig	\N	04179	Germany
10278	BERGS	8	1996-08-12	1996-09-09	1996-08-16	2	92.69	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10279	LEHMS	8	1996-08-13	1996-09-10	1996-08-16	2	25.83	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10280	BERGS	2	1996-08-14	1996-09-11	1996-09-12	1	8.98	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10281	ROMEY	4	1996-08-14	1996-08-28	1996-08-21	1	2.94	Romero y tomillo	Gran Vía, 1	Madrid	\N	28001	Spain
10282	ROMEY	4	1996-08-15	1996-09-12	1996-08-21	1	12.69	Romero y tomillo	Gran Vía, 1	Madrid	\N	28001	Spain
10283	LILAS	3	1996-08-16	1996-09-13	1996-08-23	3	84.81	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
10284	LEHMS	4	1996-08-19	1996-09-16	1996-08-27	1	76.56	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10285	QUICK	1	1996-08-20	1996-09-17	1996-08-26	2	76.83	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10286	QUICK	8	1996-08-21	1996-09-18	1996-08-30	3	229.24	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10287	RICAR	8	1996-08-22	1996-09-19	1996-08-28	3	12.76	Ricardo Adocicados	Av. Copacabana, 267	Rio de Janeiro	RJ	02389-890	Brazil
10288	REGGC	4	1996-08-23	1996-09-20	1996-09-03	1	7.45	Reggiani Caseifici	Strada Provinciale 124	Reggio Emilia	\N	42100	Italy
10289	BSBEV	7	1996-08-26	1996-09-23	1996-08-28	3	22.77	B's Beverages	Fauntleroy Circus	London	\N	EC2 5NT	UK
10290	COMMI	8	1996-08-27	1996-09-24	1996-09-03	1	79.7	Comércio Mineiro	Av. dos Lusíadas, 23	Sao Paulo	SP	05432-043	Brazil
10291	QUEDE	6	1996-08-27	1996-09-24	1996-09-04	2	6.4	Que Delícia	Rua da Panificadora, 12	Rio de Janeiro	RJ	02389-673	Brazil
10292	TRADH	1	1996-08-28	1996-09-25	1996-09-02	2	1.35	Tradiçao Hipermercados	Av. Inês de Castro, 414	Sao Paulo	SP	05634-030	Brazil
10293	TORTU	1	1996-08-29	1996-09-26	1996-09-11	3	21.18	Tortuga Restaurante	Avda. Azteca 123	México D.F.	\N	05033	Mexico
10294	RATTC	4	1996-08-30	1996-09-27	1996-09-05	2	147.26	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10295	VINET	2	1996-09-02	1996-09-30	1996-09-10	2	1.15	Vins et alcools Chevalier	59 rue de l'Abbaye	Reims	\N	51100	France
10296	LILAS	6	1996-09-03	1996-10-01	1996-09-11	1	0.12	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
10297	BLONP	5	1996-09-04	1996-10-16	1996-09-10	2	5.74	Blondel père et fils	24, place Kléber	Strasbourg	\N	67000	France
10298	HUNGO	6	1996-09-05	1996-10-03	1996-09-11	2	168.22	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10299	RICAR	4	1996-09-06	1996-10-04	1996-09-13	2	29.76	Ricardo Adocicados	Av. Copacabana, 267	Rio de Janeiro	RJ	02389-890	Brazil
10300	MAGAA	2	1996-09-09	1996-10-07	1996-09-18	2	17.68	Magazzini Alimentari Riuniti	Via Ludovico il Moro 22	Bergamo	\N	24100	Italy
10301	WANDK	8	1996-09-09	1996-10-07	1996-09-17	2	45.08	Die Wandernde Kuh	Adenauerallee 900	Stuttgart	\N	70563	Germany
10302	SUPRD	4	1996-09-10	1996-10-08	1996-10-09	2	6.27	Suprêmes délices	Boulevard Tirou, 255	Charleroi	\N	B-6000	Belgium
10303	GODOS	7	1996-09-11	1996-10-09	1996-09-18	2	107.83	Godos Cocina Típica	C/ Romero, 33	Sevilla	\N	41101	Spain
10304	TORTU	1	1996-09-12	1996-10-10	1996-09-17	2	63.79	Tortuga Restaurante	Avda. Azteca 123	México D.F.	\N	05033	Mexico
10305	OLDWO	8	1996-09-13	1996-10-11	1996-10-09	3	257.62	Old World Delicatessen	2743 Bering St.	Anchorage	AK	99508	USA
10306	ROMEY	1	1996-09-16	1996-10-14	1996-09-23	3	7.56	Romero y tomillo	Gran Vía, 1	Madrid	\N	28001	Spain
10307	LONEP	2	1996-09-17	1996-10-15	1996-09-25	2	0.56	Lonesome Pine Restaurant	89 Chiaroscuro Rd.	Portland	OR	97219	USA
10308	ANATR	7	1996-09-18	1996-10-16	1996-09-24	3	1.61	Ana Trujillo Emparedados y helados	Avda. de la Constitución 2222	México D.F.	\N	05021	Mexico
10309	HUNGO	3	1996-09-19	1996-10-17	1996-10-23	1	47.3	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10310	THEBI	8	1996-09-20	1996-10-18	1996-09-27	2	17.52	The Big Cheese	89 Jefferson Way Suite 2	Portland	OR	97201	USA
10311	DUMON	1	1996-09-20	1996-10-04	1996-09-26	3	24.69	Du monde entier	67, rue des Cinquante Otages	Nantes	\N	44000	France
10312	WANDK	2	1996-09-23	1996-10-21	1996-10-03	2	40.26	Die Wandernde Kuh	Adenauerallee 900	Stuttgart	\N	70563	Germany
10313	QUICK	2	1996-09-24	1996-10-22	1996-10-04	2	1.96	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10314	RATTC	1	1996-09-25	1996-10-23	1996-10-04	2	74.16	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10315	ISLAT	4	1996-09-26	1996-10-24	1996-10-03	2	41.76	Island Trading	Garden House Crowther Way	Cowes	Isle of Wight	PO31 7PJ	UK
10316	RATTC	1	1996-09-27	1996-10-25	1996-10-08	3	150.15	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10317	LONEP	6	1996-09-30	1996-10-28	1996-10-10	1	12.69	Lonesome Pine Restaurant	89 Chiaroscuro Rd.	Portland	OR	97219	USA
10318	ISLAT	8	1996-10-01	1996-10-29	1996-10-04	2	4.73	Island Trading	Garden House Crowther Way	Cowes	Isle of Wight	PO31 7PJ	UK
10319	TORTU	7	1996-10-02	1996-10-30	1996-10-11	3	64.5	Tortuga Restaurante	Avda. Azteca 123	México D.F.	\N	05033	Mexico
10320	WARTH	5	1996-10-03	1996-10-17	1996-10-18	3	34.57	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10321	ISLAT	3	1996-10-03	1996-10-31	1996-10-11	2	3.43	Island Trading	Garden House Crowther Way	Cowes	Isle of Wight	PO31 7PJ	UK
10322	PERIC	7	1996-10-04	1996-11-01	1996-10-23	3	0.4	Pericles Comidas clásicas	Calle Dr. Jorge Cash 321	México D.F.	\N	05033	Mexico
10323	KOENE	4	1996-10-07	1996-11-04	1996-10-14	1	4.88	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
10324	SAVEA	9	1996-10-08	1996-11-05	1996-10-10	1	214.27	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10325	KOENE	1	1996-10-09	1996-10-23	1996-10-14	3	64.86	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
10326	BOLID	4	1996-10-10	1996-11-07	1996-10-14	2	77.92	Bólido Comidas preparadas	C/ Araquil, 67	Madrid	\N	28023	Spain
10327	FOLKO	2	1996-10-11	1996-11-08	1996-10-14	1	63.36	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10328	FURIB	4	1996-10-14	1996-11-11	1996-10-17	3	87.03	Furia Bacalhau e Frutos do Mar	Jardim das rosas n. 32	Lisboa	\N	1675	Portugal
10329	SPLIR	4	1996-10-15	1996-11-26	1996-10-23	2	191.67	Split Rail Beer & Ale	P.O. Box 555	Lander	WY	82520	USA
10330	LILAS	3	1996-10-16	1996-11-13	1996-10-28	1	12.75	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
10331	BONAP	9	1996-10-16	1996-11-27	1996-10-21	1	10.19	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10332	MEREP	3	1996-10-17	1996-11-28	1996-10-21	2	52.84	Mère Paillarde	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada
10333	WARTH	5	1996-10-18	1996-11-15	1996-10-25	3	0.59	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10334	VICTE	8	1996-10-21	1996-11-18	1996-10-28	2	8.56	Victuailles en stock	2, rue du Commerce	Lyon	\N	69004	France
10335	HUNGO	7	1996-10-22	1996-11-19	1996-10-24	2	42.11	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10336	PRINI	7	1996-10-23	1996-11-20	1996-10-25	2	15.51	Princesa Isabel Vinhos	Estrada da saúde n. 58	Lisboa	\N	1756	Portugal
10337	FRANK	4	1996-10-24	1996-11-21	1996-10-29	3	108.26	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10338	OLDWO	4	1996-10-25	1996-11-22	1996-10-29	3	84.21	Old World Delicatessen	2743 Bering St.	Anchorage	AK	99508	USA
10339	MEREP	2	1996-10-28	1996-11-25	1996-11-04	2	15.66	Mère Paillarde	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada
10340	BONAP	1	1996-10-29	1996-11-26	1996-11-08	3	166.31	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10341	SIMOB	7	1996-10-29	1996-11-26	1996-11-05	3	26.78	Simons bistro	Vinbæltet 34	Kobenhavn	\N	1734	Denmark
10342	FRANK	4	1996-10-30	1996-11-13	1996-11-04	2	54.83	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10343	LEHMS	4	1996-10-31	1996-11-28	1996-11-06	1	110.37	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10344	WHITC	4	1996-11-01	1996-11-29	1996-11-05	2	23.29	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
10345	QUICK	2	1996-11-04	1996-12-02	1996-11-11	2	249.06	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10346	RATTC	3	1996-11-05	1996-12-17	1996-11-08	3	142.08	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10347	FAMIA	4	1996-11-06	1996-12-04	1996-11-08	3	3.1	Familia Arquibaldo	Rua Orós, 92	Sao Paulo	SP	05442-030	Brazil
10348	WANDK	4	1996-11-07	1996-12-05	1996-11-15	2	0.78	Die Wandernde Kuh	Adenauerallee 900	Stuttgart	\N	70563	Germany
10349	SPLIR	7	1996-11-08	1996-12-06	1996-11-15	1	8.63	Split Rail Beer & Ale	P.O. Box 555	Lander	WY	82520	USA
10350	LAMAI	6	1996-11-11	1996-12-09	1996-12-03	2	64.19	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
10351	ERNSH	1	1996-11-11	1996-12-09	1996-11-20	1	162.33	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10352	FURIB	3	1996-11-12	1996-11-26	1996-11-18	3	1.3	Furia Bacalhau e Frutos do Mar	Jardim das rosas n. 32	Lisboa	\N	1675	Portugal
10353	PICCO	7	1996-11-13	1996-12-11	1996-11-25	3	360.63	Piccolo und mehr	Geislweg 14	Salzburg	\N	5020	Austria
10354	PERIC	8	1996-11-14	1996-12-12	1996-11-20	3	53.8	Pericles Comidas clásicas	Calle Dr. Jorge Cash 321	México D.F.	\N	05033	Mexico
10355	AROUT	6	1996-11-15	1996-12-13	1996-11-20	1	41.95	Around the Horn	Brook Farm Stratford St. Mary	Colchester	Essex	CO7 6JX	UK
10356	WANDK	6	1996-11-18	1996-12-16	1996-11-27	2	36.71	Die Wandernde Kuh	Adenauerallee 900	Stuttgart	\N	70563	Germany
10357	LILAS	1	1996-11-19	1996-12-17	1996-12-02	3	34.88	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
10358	LAMAI	5	1996-11-20	1996-12-18	1996-11-27	1	19.64	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
10359	SEVES	5	1996-11-21	1996-12-19	1996-11-26	3	288.43	Seven Seas Imports	90 Wadhurst Rd.	London	\N	OX15 4NB	UK
10360	BLONP	4	1996-11-22	1996-12-20	1996-12-02	3	131.7	Blondel père et fils	24, place Kléber	Strasbourg	\N	67000	France
10361	QUICK	1	1996-11-22	1996-12-20	1996-12-03	2	183.17	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10362	BONAP	3	1996-11-25	1996-12-23	1996-11-28	1	96.04	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10363	DRACD	4	1996-11-26	1996-12-24	1996-12-04	3	30.54	Drachenblut Delikatessen	Walserweg 21	Aachen	\N	52066	Germany
10364	EASTC	1	1996-11-26	1997-01-07	1996-12-04	1	71.97	Eastern Connection	35 King George	London	\N	WX3 6FW	UK
10365	ANTON	3	1996-11-27	1996-12-25	1996-12-02	2	22	Antonio Moreno Taquería	Mataderos  2312	México D.F.	\N	05023	Mexico
10366	GALED	8	1996-11-28	1997-01-09	1996-12-30	2	10.14	Galería del gastronómo	Rambla de Cataluña, 23	Barcelona	\N	8022	Spain
10367	VAFFE	7	1996-11-28	1996-12-26	1996-12-02	3	13.55	Vaffeljernet	Smagsloget 45	Århus	\N	8200	Denmark
10368	ERNSH	2	1996-11-29	1996-12-27	1996-12-02	2	101.95	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10369	SPLIR	8	1996-12-02	1996-12-30	1996-12-09	2	195.68	Split Rail Beer & Ale	P.O. Box 555	Lander	WY	82520	USA
10370	CHOPS	6	1996-12-03	1996-12-31	1996-12-27	2	1.17	Chop-suey Chinese	Hauptstr. 31	Bern	\N	3012	Switzerland
10371	LAMAI	1	1996-12-03	1996-12-31	1996-12-24	1	0.45	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
10372	QUEEN	5	1996-12-04	1997-01-01	1996-12-09	2	890.78	Queen Cozinha	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil
10373	HUNGO	4	1996-12-05	1997-01-02	1996-12-11	3	124.12	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10374	WOLZA	1	1996-12-05	1997-01-02	1996-12-09	3	3.94	Wolski Zajazd	ul. Filtrowa 68	Warszawa	\N	01-012	Poland
10375	HUNGC	3	1996-12-06	1997-01-03	1996-12-09	2	20.12	Hungry Coyote Import Store	City Center Plaza 516 Main St.	Elgin	OR	97827	USA
10376	MEREP	1	1996-12-09	1997-01-06	1996-12-13	2	20.39	Mère Paillarde	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada
10377	SEVES	1	1996-12-09	1997-01-06	1996-12-13	3	22.21	Seven Seas Imports	90 Wadhurst Rd.	London	\N	OX15 4NB	UK
10378	FOLKO	5	1996-12-10	1997-01-07	1996-12-19	3	5.44	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10379	QUEDE	2	1996-12-11	1997-01-08	1996-12-13	1	45.03	Que Delícia	Rua da Panificadora, 12	Rio de Janeiro	RJ	02389-673	Brazil
10380	HUNGO	8	1996-12-12	1997-01-09	1997-01-16	3	35.03	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10381	LILAS	3	1996-12-12	1997-01-09	1996-12-13	3	7.99	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
10382	ERNSH	4	1996-12-13	1997-01-10	1996-12-16	1	94.77	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10383	AROUT	8	1996-12-16	1997-01-13	1996-12-18	3	34.24	Around the Horn	Brook Farm Stratford St. Mary	Colchester	Essex	CO7 6JX	UK
10384	BERGS	3	1996-12-16	1997-01-13	1996-12-20	3	168.64	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10385	SPLIR	1	1996-12-17	1997-01-14	1996-12-23	2	30.96	Split Rail Beer & Ale	P.O. Box 555	Lander	WY	82520	USA
10386	FAMIA	9	1996-12-18	1997-01-01	1996-12-25	3	13.99	Familia Arquibaldo	Rua Orós, 92	Sao Paulo	SP	05442-030	Brazil
10387	SANTG	1	1996-12-18	1997-01-15	1996-12-20	2	93.63	Santé Gourmet	Erling Skakkes gate 78	Stavern	\N	4110	Norway
10388	SEVES	2	1996-12-19	1997-01-16	1996-12-20	1	34.86	Seven Seas Imports	90 Wadhurst Rd.	London	\N	OX15 4NB	UK
10389	BOTTM	4	1996-12-20	1997-01-17	1996-12-24	2	47.42	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
10390	ERNSH	6	1996-12-23	1997-01-20	1996-12-26	1	126.38	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10391	DRACD	3	1996-12-23	1997-01-20	1996-12-31	3	5.45	Drachenblut Delikatessen	Walserweg 21	Aachen	\N	52066	Germany
10392	PICCO	2	1996-12-24	1997-01-21	1997-01-01	3	122.46	Piccolo und mehr	Geislweg 14	Salzburg	\N	5020	Austria
10393	SAVEA	1	1996-12-25	1997-01-22	1997-01-03	3	126.56	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10394	HUNGC	1	1996-12-25	1997-01-22	1997-01-03	3	30.34	Hungry Coyote Import Store	City Center Plaza 516 Main St.	Elgin	OR	97827	USA
10395	HILAA	6	1996-12-26	1997-01-23	1997-01-03	1	184.41	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10396	FRANK	1	1996-12-27	1997-01-10	1997-01-06	3	135.35	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10397	PRINI	5	1996-12-27	1997-01-24	1997-01-02	1	60.26	Princesa Isabel Vinhos	Estrada da saúde n. 58	Lisboa	\N	1756	Portugal
10398	SAVEA	2	1996-12-30	1997-01-27	1997-01-09	3	89.16	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10399	VAFFE	8	1996-12-31	1997-01-14	1997-01-08	3	27.36	Vaffeljernet	Smagsloget 45	Århus	\N	8200	Denmark
10400	EASTC	1	1997-01-01	1997-01-29	1997-01-16	3	83.93	Eastern Connection	35 King George	London	\N	WX3 6FW	UK
10401	RATTC	1	1997-01-01	1997-01-29	1997-01-10	1	12.51	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10402	ERNSH	8	1997-01-02	1997-02-13	1997-01-10	2	67.88	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10403	ERNSH	4	1997-01-03	1997-01-31	1997-01-09	3	73.79	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10404	MAGAA	2	1997-01-03	1997-01-31	1997-01-08	1	155.97	Magazzini Alimentari Riuniti	Via Ludovico il Moro 22	Bergamo	\N	24100	Italy
10405	LINOD	1	1997-01-06	1997-02-03	1997-01-22	1	34.82	LINO-Delicateses	Ave. 5 de Mayo Porlamar	I. de Margarita	Nueva Esparta	4980	Venezuela
10406	QUEEN	7	1997-01-07	1997-02-18	1997-01-13	1	108.04	Queen Cozinha	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil
10407	OTTIK	2	1997-01-07	1997-02-04	1997-01-30	2	91.48	Ottilies Käseladen	Mehrheimerstr. 369	Köln	\N	50739	Germany
10408	FOLIG	8	1997-01-08	1997-02-05	1997-01-14	1	11.26	Folies gourmandes	184, chaussée de Tournai	Lille	\N	59000	France
10409	OCEAN	3	1997-01-09	1997-02-06	1997-01-14	1	29.83	Océano Atlántico Ltda.	Ing. Gustavo Moncada 8585 Piso 20-A	Buenos Aires	\N	1010	Argentina
10410	BOTTM	3	1997-01-10	1997-02-07	1997-01-15	3	2.4	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
10411	BOTTM	9	1997-01-10	1997-02-07	1997-01-21	3	23.65	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
10412	WARTH	8	1997-01-13	1997-02-10	1997-01-15	2	3.77	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10413	LAMAI	3	1997-01-14	1997-02-11	1997-01-16	2	95.66	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
10414	FAMIA	2	1997-01-14	1997-02-11	1997-01-17	3	21.48	Familia Arquibaldo	Rua Orós, 92	Sao Paulo	SP	05442-030	Brazil
10415	HUNGC	3	1997-01-15	1997-02-12	1997-01-24	1	0.2	Hungry Coyote Import Store	City Center Plaza 516 Main St.	Elgin	OR	97827	USA
10416	WARTH	8	1997-01-16	1997-02-13	1997-01-27	3	22.72	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10417	SIMOB	4	1997-01-16	1997-02-13	1997-01-28	3	70.29	Simons bistro	Vinbæltet 34	Kobenhavn	\N	1734	Denmark
10418	QUICK	4	1997-01-17	1997-02-14	1997-01-24	1	17.55	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10419	RICSU	4	1997-01-20	1997-02-17	1997-01-30	2	137.35	Richter Supermarkt	Starenweg 5	Genève	\N	1204	Switzerland
10420	WELLI	3	1997-01-21	1997-02-18	1997-01-27	1	44.12	Wellington Importadora	Rua do Mercado, 12	Resende	SP	08737-363	Brazil
10421	QUEDE	8	1997-01-21	1997-03-04	1997-01-27	1	99.23	Que Delícia	Rua da Panificadora, 12	Rio de Janeiro	RJ	02389-673	Brazil
10422	FRANS	2	1997-01-22	1997-02-19	1997-01-31	1	3.02	Franchi S.p.A.	Via Monte Bianco 34	Torino	\N	10100	Italy
10423	GOURL	6	1997-01-23	1997-02-06	1997-02-24	3	24.5	Gourmet Lanchonetes	Av. Brasil, 442	Campinas	SP	04876-786	Brazil
10424	MEREP	7	1997-01-23	1997-02-20	1997-01-27	2	370.61	Mère Paillarde	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada
10425	LAMAI	6	1997-01-24	1997-02-21	1997-02-14	2	7.93	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
10426	GALED	4	1997-01-27	1997-02-24	1997-02-06	1	18.69	Galería del gastronómo	Rambla de Cataluña, 23	Barcelona	\N	8022	Spain
10427	PICCO	4	1997-01-27	1997-02-24	1997-03-03	2	31.29	Piccolo und mehr	Geislweg 14	Salzburg	\N	5020	Austria
10428	REGGC	7	1997-01-28	1997-02-25	1997-02-04	1	11.09	Reggiani Caseifici	Strada Provinciale 124	Reggio Emilia	\N	42100	Italy
10429	HUNGO	3	1997-01-29	1997-03-12	1997-02-07	2	56.63	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10430	ERNSH	4	1997-01-30	1997-02-13	1997-02-03	1	458.78	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10431	BOTTM	4	1997-01-30	1997-02-13	1997-02-07	2	44.17	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
10432	SPLIR	3	1997-01-31	1997-02-14	1997-02-07	2	4.34	Split Rail Beer & Ale	P.O. Box 555	Lander	WY	82520	USA
10433	PRINI	3	1997-02-03	1997-03-03	1997-03-04	3	73.83	Princesa Isabel Vinhos	Estrada da saúde n. 58	Lisboa	\N	1756	Portugal
10434	FOLKO	3	1997-02-03	1997-03-03	1997-02-13	2	17.92	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10435	CONSH	8	1997-02-04	1997-03-18	1997-02-07	2	9.21	Consolidated Holdings	Berkeley Gardens 12  Brewery	London	\N	WX1 6LT	UK
10436	BLONP	3	1997-02-05	1997-03-05	1997-02-11	2	156.66	Blondel père et fils	24, place Kléber	Strasbourg	\N	67000	France
10437	WARTH	8	1997-02-05	1997-03-05	1997-02-12	1	19.97	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10438	TOMSP	3	1997-02-06	1997-03-06	1997-02-14	2	8.24	Toms Spezialitäten	Luisenstr. 48	Münster	\N	44087	Germany
10439	MEREP	6	1997-02-07	1997-03-07	1997-02-10	3	4.07	Mère Paillarde	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada
10440	SAVEA	4	1997-02-10	1997-03-10	1997-02-28	2	86.53	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10441	OLDWO	3	1997-02-10	1997-03-24	1997-03-14	2	73.02	Old World Delicatessen	2743 Bering St.	Anchorage	AK	99508	USA
10442	ERNSH	3	1997-02-11	1997-03-11	1997-02-18	2	47.94	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10443	REGGC	8	1997-02-12	1997-03-12	1997-02-14	1	13.95	Reggiani Caseifici	Strada Provinciale 124	Reggio Emilia	\N	42100	Italy
10444	BERGS	3	1997-02-12	1997-03-12	1997-02-21	3	3.5	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10445	BERGS	3	1997-02-13	1997-03-13	1997-02-20	1	9.3	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10446	TOMSP	6	1997-02-14	1997-03-14	1997-02-19	1	14.68	Toms Spezialitäten	Luisenstr. 48	Münster	\N	44087	Germany
10447	RICAR	4	1997-02-14	1997-03-14	1997-03-07	2	68.66	Ricardo Adocicados	Av. Copacabana, 267	Rio de Janeiro	RJ	02389-890	Brazil
10448	RANCH	4	1997-02-17	1997-03-17	1997-02-24	2	38.82	Rancho grande	Av. del Libertador 900	Buenos Aires	\N	1010	Argentina
10449	BLONP	3	1997-02-18	1997-03-18	1997-02-27	2	53.3	Blondel père et fils	24, place Kléber	Strasbourg	\N	67000	France
10450	VICTE	8	1997-02-19	1997-03-19	1997-03-11	2	7.23	Victuailles en stock	2, rue du Commerce	Lyon	\N	69004	France
10451	QUICK	4	1997-02-19	1997-03-05	1997-03-12	3	189.09	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10452	SAVEA	8	1997-02-20	1997-03-20	1997-02-26	1	140.26	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10453	AROUT	1	1997-02-21	1997-03-21	1997-02-26	2	25.36	Around the Horn	Brook Farm Stratford St. Mary	Colchester	Essex	CO7 6JX	UK
10454	LAMAI	4	1997-02-21	1997-03-21	1997-02-25	3	2.74	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
10455	WARTH	8	1997-02-24	1997-04-07	1997-03-03	2	180.45	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10456	KOENE	8	1997-02-25	1997-04-08	1997-02-28	2	8.12	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
10457	KOENE	2	1997-02-25	1997-03-25	1997-03-03	1	11.57	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
10458	SUPRD	7	1997-02-26	1997-03-26	1997-03-04	3	147.06	Suprêmes délices	Boulevard Tirou, 255	Charleroi	\N	B-6000	Belgium
10459	VICTE	4	1997-02-27	1997-03-27	1997-02-28	2	25.09	Victuailles en stock	2, rue du Commerce	Lyon	\N	69004	France
10460	FOLKO	8	1997-02-28	1997-03-28	1997-03-03	1	16.27	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10461	LILAS	1	1997-02-28	1997-03-28	1997-03-05	3	148.61	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
10462	CONSH	2	1997-03-03	1997-03-31	1997-03-18	1	6.17	Consolidated Holdings	Berkeley Gardens 12  Brewery	London	\N	WX1 6LT	UK
10463	SUPRD	5	1997-03-04	1997-04-01	1997-03-06	3	14.78	Suprêmes délices	Boulevard Tirou, 255	Charleroi	\N	B-6000	Belgium
10464	FURIB	4	1997-03-04	1997-04-01	1997-03-14	2	89	Furia Bacalhau e Frutos do Mar	Jardim das rosas n. 32	Lisboa	\N	1675	Portugal
10465	VAFFE	1	1997-03-05	1997-04-02	1997-03-14	3	145.04	Vaffeljernet	Smagsloget 45	Århus	\N	8200	Denmark
10466	COMMI	4	1997-03-06	1997-04-03	1997-03-13	1	11.93	Comércio Mineiro	Av. dos Lusíadas, 23	Sao Paulo	SP	05432-043	Brazil
10467	MAGAA	8	1997-03-06	1997-04-03	1997-03-11	2	4.93	Magazzini Alimentari Riuniti	Via Ludovico il Moro 22	Bergamo	\N	24100	Italy
10468	KOENE	3	1997-03-07	1997-04-04	1997-03-12	3	44.12	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
10469	WHITC	1	1997-03-10	1997-04-07	1997-03-14	1	60.18	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
10470	BONAP	4	1997-03-11	1997-04-08	1997-03-14	2	64.56	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10471	BSBEV	2	1997-03-11	1997-04-08	1997-03-18	3	45.59	B's Beverages	Fauntleroy Circus	London	\N	EC2 5NT	UK
10472	SEVES	8	1997-03-12	1997-04-09	1997-03-19	1	4.2	Seven Seas Imports	90 Wadhurst Rd.	London	\N	OX15 4NB	UK
10473	ISLAT	1	1997-03-13	1997-03-27	1997-03-21	3	16.37	Island Trading	Garden House Crowther Way	Cowes	Isle of Wight	PO31 7PJ	UK
10474	PERIC	5	1997-03-13	1997-04-10	1997-03-21	2	83.49	Pericles Comidas clásicas	Calle Dr. Jorge Cash 321	México D.F.	\N	05033	Mexico
10475	SUPRD	9	1997-03-14	1997-04-11	1997-04-04	1	68.52	Suprêmes délices	Boulevard Tirou, 255	Charleroi	\N	B-6000	Belgium
10476	HILAA	8	1997-03-17	1997-04-14	1997-03-24	3	4.41	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10477	PRINI	5	1997-03-17	1997-04-14	1997-03-25	2	13.02	Princesa Isabel Vinhos	Estrada da saúde n. 58	Lisboa	\N	1756	Portugal
10478	VICTE	2	1997-03-18	1997-04-01	1997-03-26	3	4.81	Victuailles en stock	2, rue du Commerce	Lyon	\N	69004	France
10479	RATTC	3	1997-03-19	1997-04-16	1997-03-21	3	708.95	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10480	FOLIG	6	1997-03-20	1997-04-17	1997-03-24	2	1.35	Folies gourmandes	184, chaussée de Tournai	Lille	\N	59000	France
10481	RICAR	8	1997-03-20	1997-04-17	1997-03-25	2	64.33	Ricardo Adocicados	Av. Copacabana, 267	Rio de Janeiro	RJ	02389-890	Brazil
10482	LAZYK	1	1997-03-21	1997-04-18	1997-04-10	3	7.48	Lazy K Kountry Store	12 Orchestra Terrace	Walla Walla	WA	99362	USA
10483	WHITC	7	1997-03-24	1997-04-21	1997-04-25	2	15.28	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
10484	BSBEV	3	1997-03-24	1997-04-21	1997-04-01	3	6.88	B's Beverages	Fauntleroy Circus	London	\N	EC2 5NT	UK
10485	LINOD	4	1997-03-25	1997-04-08	1997-03-31	2	64.45	LINO-Delicateses	Ave. 5 de Mayo Porlamar	I. de Margarita	Nueva Esparta	4980	Venezuela
10486	HILAA	1	1997-03-26	1997-04-23	1997-04-02	2	30.53	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10487	QUEEN	2	1997-03-26	1997-04-23	1997-03-28	2	71.07	Queen Cozinha	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil
10488	FRANK	8	1997-03-27	1997-04-24	1997-04-02	2	4.93	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10489	PICCO	6	1997-03-28	1997-04-25	1997-04-09	2	5.29	Piccolo und mehr	Geislweg 14	Salzburg	\N	5020	Austria
10490	HILAA	7	1997-03-31	1997-04-28	1997-04-03	2	210.19	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10491	FURIB	8	1997-03-31	1997-04-28	1997-04-08	3	16.96	Furia Bacalhau e Frutos do Mar	Jardim das rosas n. 32	Lisboa	\N	1675	Portugal
10492	BOTTM	3	1997-04-01	1997-04-29	1997-04-11	1	62.89	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
10493	LAMAI	4	1997-04-02	1997-04-30	1997-04-10	3	10.64	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
10494	COMMI	4	1997-04-02	1997-04-30	1997-04-09	2	65.99	Comércio Mineiro	Av. dos Lusíadas, 23	Sao Paulo	SP	05432-043	Brazil
10495	LAUGB	3	1997-04-03	1997-05-01	1997-04-11	3	4.65	Laughing Bacchus Wine Cellars	2319 Elm St.	Vancouver	BC	V3F 2K1	Canada
10496	TRADH	7	1997-04-04	1997-05-02	1997-04-07	2	46.77	Tradiçao Hipermercados	Av. Inês de Castro, 414	Sao Paulo	SP	05634-030	Brazil
10497	LEHMS	7	1997-04-04	1997-05-02	1997-04-07	1	36.21	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10498	HILAA	8	1997-04-07	1997-05-05	1997-04-11	2	29.75	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10499	LILAS	4	1997-04-08	1997-05-06	1997-04-16	2	102.02	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
10500	LAMAI	6	1997-04-09	1997-05-07	1997-04-17	1	42.68	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
10501	BLAUS	9	1997-04-09	1997-05-07	1997-04-16	3	8.85	Blauer See Delikatessen	Forsterstr. 57	Mannheim	\N	68306	Germany
10502	PERIC	2	1997-04-10	1997-05-08	1997-04-29	1	69.32	Pericles Comidas clásicas	Calle Dr. Jorge Cash 321	México D.F.	\N	05033	Mexico
10503	HUNGO	6	1997-04-11	1997-05-09	1997-04-16	2	16.74	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10504	WHITC	4	1997-04-11	1997-05-09	1997-04-18	3	59.13	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
10505	MEREP	3	1997-04-14	1997-05-12	1997-04-21	3	7.13	Mère Paillarde	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada
10506	KOENE	9	1997-04-15	1997-05-13	1997-05-02	2	21.19	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
10507	ANTON	7	1997-04-15	1997-05-13	1997-04-22	1	47.45	Antonio Moreno Taquería	Mataderos  2312	México D.F.	\N	05023	Mexico
10508	OTTIK	1	1997-04-16	1997-05-14	1997-05-13	2	4.99	Ottilies Käseladen	Mehrheimerstr. 369	Köln	\N	50739	Germany
10509	BLAUS	4	1997-04-17	1997-05-15	1997-04-29	1	0.15	Blauer See Delikatessen	Forsterstr. 57	Mannheim	\N	68306	Germany
10510	SAVEA	6	1997-04-18	1997-05-16	1997-04-28	3	367.63	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10511	BONAP	4	1997-04-18	1997-05-16	1997-04-21	3	350.64	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10512	FAMIA	7	1997-04-21	1997-05-19	1997-04-24	2	3.53	Familia Arquibaldo	Rua Orós, 92	Sao Paulo	SP	05442-030	Brazil
10513	WANDK	7	1997-04-22	1997-06-03	1997-04-28	1	105.65	Die Wandernde Kuh	Adenauerallee 900	Stuttgart	\N	70563	Germany
10514	ERNSH	3	1997-04-22	1997-05-20	1997-05-16	2	789.95	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10515	QUICK	2	1997-04-23	1997-05-07	1997-05-23	1	204.47	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10516	HUNGO	2	1997-04-24	1997-05-22	1997-05-01	3	62.78	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10517	NORTS	3	1997-04-24	1997-05-22	1997-04-29	3	32.07	North/South	South House 300 Queensbridge	London	\N	SW7 1RZ	UK
10518	TORTU	4	1997-04-25	1997-05-09	1997-05-05	2	218.15	Tortuga Restaurante	Avda. Azteca 123	México D.F.	\N	05033	Mexico
10519	CHOPS	6	1997-04-28	1997-05-26	1997-05-01	3	91.76	Chop-suey Chinese	Hauptstr. 31	Bern	\N	3012	Switzerland
10520	SANTG	7	1997-04-29	1997-05-27	1997-05-01	1	13.37	Santé Gourmet	Erling Skakkes gate 78	Stavern	\N	4110	Norway
10521	CACTU	8	1997-04-29	1997-05-27	1997-05-02	2	17.22	Cactus Comidas para llevar	Cerrito 333	Buenos Aires	\N	1010	Argentina
10522	LEHMS	4	1997-04-30	1997-05-28	1997-05-06	1	45.33	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10523	SEVES	7	1997-05-01	1997-05-29	1997-05-30	2	77.63	Seven Seas Imports	90 Wadhurst Rd.	London	\N	OX15 4NB	UK
10524	BERGS	1	1997-05-01	1997-05-29	1997-05-07	2	244.79	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10525	BONAP	1	1997-05-02	1997-05-30	1997-05-23	2	11.06	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10526	WARTH	4	1997-05-05	1997-06-02	1997-05-15	2	58.59	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10527	QUICK	7	1997-05-05	1997-06-02	1997-05-07	1	41.9	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10528	GREAL	6	1997-05-06	1997-05-20	1997-05-09	2	3.35	Great Lakes Food Market	2732 Baker Blvd.	Eugene	OR	97403	USA
10529	MAISD	5	1997-05-07	1997-06-04	1997-05-09	2	66.69	Maison Dewey	Rue Joseph-Bens 532	Bruxelles	\N	B-1180	Belgium
10530	PICCO	3	1997-05-08	1997-06-05	1997-05-12	2	339.22	Piccolo und mehr	Geislweg 14	Salzburg	\N	5020	Austria
10531	OCEAN	7	1997-05-08	1997-06-05	1997-05-19	1	8.12	Océano Atlántico Ltda.	Ing. Gustavo Moncada 8585 Piso 20-A	Buenos Aires	\N	1010	Argentina
10532	EASTC	7	1997-05-09	1997-06-06	1997-05-12	3	74.46	Eastern Connection	35 King George	London	\N	WX3 6FW	UK
10533	FOLKO	8	1997-05-12	1997-06-09	1997-05-22	1	188.04	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10534	LEHMS	8	1997-05-12	1997-06-09	1997-05-14	2	27.94	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10535	ANTON	4	1997-05-13	1997-06-10	1997-05-21	1	15.64	Antonio Moreno Taquería	Mataderos  2312	México D.F.	\N	05023	Mexico
10536	LEHMS	3	1997-05-14	1997-06-11	1997-06-06	2	58.88	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10537	RICSU	1	1997-05-14	1997-05-28	1997-05-19	1	78.85	Richter Supermarkt	Starenweg 5	Genève	\N	1204	Switzerland
10538	BSBEV	9	1997-05-15	1997-06-12	1997-05-16	3	4.87	B's Beverages	Fauntleroy Circus	London	\N	EC2 5NT	UK
10539	BSBEV	6	1997-05-16	1997-06-13	1997-05-23	3	12.36	B's Beverages	Fauntleroy Circus	London	\N	EC2 5NT	UK
10540	QUICK	3	1997-05-19	1997-06-16	1997-06-13	3	1007.64	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10541	HANAR	2	1997-05-19	1997-06-16	1997-05-29	1	68.65	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
10542	KOENE	1	1997-05-20	1997-06-17	1997-05-26	3	10.95	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
10543	LILAS	8	1997-05-21	1997-06-18	1997-05-23	2	48.17	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
10544	LONEP	4	1997-05-21	1997-06-18	1997-05-30	1	24.91	Lonesome Pine Restaurant	89 Chiaroscuro Rd.	Portland	OR	97219	USA
10545	LAZYK	8	1997-05-22	1997-06-19	1997-06-26	2	11.92	Lazy K Kountry Store	12 Orchestra Terrace	Walla Walla	WA	99362	USA
10546	VICTE	1	1997-05-23	1997-06-20	1997-05-27	3	194.72	Victuailles en stock	2, rue du Commerce	Lyon	\N	69004	France
10547	SEVES	3	1997-05-23	1997-06-20	1997-06-02	2	178.43	Seven Seas Imports	90 Wadhurst Rd.	London	\N	OX15 4NB	UK
10548	TOMSP	3	1997-05-26	1997-06-23	1997-06-02	2	1.43	Toms Spezialitäten	Luisenstr. 48	Münster	\N	44087	Germany
10549	QUICK	5	1997-05-27	1997-06-10	1997-05-30	1	171.24	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10550	GODOS	7	1997-05-28	1997-06-25	1997-06-06	3	4.32	Godos Cocina Típica	C/ Romero, 33	Sevilla	\N	41101	Spain
10551	FURIB	4	1997-05-28	1997-07-09	1997-06-06	3	72.95	Furia Bacalhau e Frutos do Mar	Jardim das rosas n. 32	Lisboa	\N	1675	Portugal
10552	HILAA	2	1997-05-29	1997-06-26	1997-06-05	1	83.22	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10553	WARTH	2	1997-05-30	1997-06-27	1997-06-03	2	149.49	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10554	OTTIK	4	1997-05-30	1997-06-27	1997-06-05	3	120.97	Ottilies Käseladen	Mehrheimerstr. 369	Köln	\N	50739	Germany
10555	SAVEA	6	1997-06-02	1997-06-30	1997-06-04	3	252.49	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10556	SIMOB	2	1997-06-03	1997-07-15	1997-06-13	1	9.8	Simons bistro	Vinbæltet 34	Kobenhavn	\N	1734	Denmark
10557	LEHMS	9	1997-06-03	1997-06-17	1997-06-06	2	96.72	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10558	AROUT	1	1997-06-04	1997-07-02	1997-06-10	2	72.97	Around the Horn	Brook Farm Stratford St. Mary	Colchester	Essex	CO7 6JX	UK
10559	BLONP	6	1997-06-05	1997-07-03	1997-06-13	1	8.05	Blondel père et fils	24, place Kléber	Strasbourg	\N	67000	France
10560	FRANK	8	1997-06-06	1997-07-04	1997-06-09	1	36.65	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10561	FOLKO	2	1997-06-06	1997-07-04	1997-06-09	2	242.21	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10562	REGGC	1	1997-06-09	1997-07-07	1997-06-12	1	22.95	Reggiani Caseifici	Strada Provinciale 124	Reggio Emilia	\N	42100	Italy
10563	RICAR	2	1997-06-10	1997-07-22	1997-06-24	2	60.43	Ricardo Adocicados	Av. Copacabana, 267	Rio de Janeiro	RJ	02389-890	Brazil
10564	RATTC	4	1997-06-10	1997-07-08	1997-06-16	3	13.75	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10565	MEREP	8	1997-06-11	1997-07-09	1997-06-18	2	7.15	Mère Paillarde	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada
10566	BLONP	9	1997-06-12	1997-07-10	1997-06-18	1	88.4	Blondel père et fils	24, place Kléber	Strasbourg	\N	67000	France
10567	HUNGO	1	1997-06-12	1997-07-10	1997-06-17	1	33.97	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10568	GALED	3	1997-06-13	1997-07-11	1997-07-09	3	6.54	Galería del gastronómo	Rambla de Cataluña, 23	Barcelona	\N	8022	Spain
10569	RATTC	5	1997-06-16	1997-07-14	1997-07-11	1	58.98	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10570	MEREP	3	1997-06-17	1997-07-15	1997-06-19	3	188.99	Mère Paillarde	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada
10571	ERNSH	8	1997-06-17	1997-07-29	1997-07-04	3	26.06	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10572	BERGS	3	1997-06-18	1997-07-16	1997-06-25	2	116.43	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10573	ANTON	7	1997-06-19	1997-07-17	1997-06-20	3	84.84	Antonio Moreno Taquería	Mataderos  2312	México D.F.	\N	05023	Mexico
10574	TRAIH	4	1997-06-19	1997-07-17	1997-06-30	2	37.6	Trail's Head Gourmet Provisioners	722 DaVinci Blvd.	Kirkland	WA	98034	USA
10575	MORGK	5	1997-06-20	1997-07-04	1997-06-30	1	127.34	Morgenstern Gesundkost	Heerstr. 22	Leipzig	\N	04179	Germany
10576	TORTU	3	1997-06-23	1997-07-07	1997-06-30	3	18.56	Tortuga Restaurante	Avda. Azteca 123	México D.F.	\N	05033	Mexico
10577	TRAIH	9	1997-06-23	1997-08-04	1997-06-30	2	25.41	Trail's Head Gourmet Provisioners	722 DaVinci Blvd.	Kirkland	WA	98034	USA
10578	BSBEV	4	1997-06-24	1997-07-22	1997-07-25	3	29.6	B's Beverages	Fauntleroy Circus	London	\N	EC2 5NT	UK
10579	LETSS	1	1997-06-25	1997-07-23	1997-07-04	2	13.73	Let's Stop N Shop	87 Polk St. Suite 5	San Francisco	CA	94117	USA
10580	OTTIK	4	1997-06-26	1997-07-24	1997-07-01	3	75.89	Ottilies Käseladen	Mehrheimerstr. 369	Köln	\N	50739	Germany
10581	FAMIA	3	1997-06-26	1997-07-24	1997-07-02	1	3.01	Familia Arquibaldo	Rua Orós, 92	Sao Paulo	SP	05442-030	Brazil
10582	BLAUS	3	1997-06-27	1997-07-25	1997-07-14	2	27.71	Blauer See Delikatessen	Forsterstr. 57	Mannheim	\N	68306	Germany
10583	WARTH	2	1997-06-30	1997-07-28	1997-07-04	2	7.28	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10584	BLONP	4	1997-06-30	1997-07-28	1997-07-04	1	59.14	Blondel père et fils	24, place Kléber	Strasbourg	\N	67000	France
10585	WELLI	7	1997-07-01	1997-07-29	1997-07-10	1	13.41	Wellington Importadora	Rua do Mercado, 12	Resende	SP	08737-363	Brazil
10586	REGGC	9	1997-07-02	1997-07-30	1997-07-09	1	0.48	Reggiani Caseifici	Strada Provinciale 124	Reggio Emilia	\N	42100	Italy
10587	QUEDE	1	1997-07-02	1997-07-30	1997-07-09	1	62.52	Que Delícia	Rua da Panificadora, 12	Rio de Janeiro	RJ	02389-673	Brazil
10588	QUICK	2	1997-07-03	1997-07-31	1997-07-10	3	194.67	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10589	GREAL	8	1997-07-04	1997-08-01	1997-07-14	2	4.42	Great Lakes Food Market	2732 Baker Blvd.	Eugene	OR	97403	USA
10590	MEREP	4	1997-07-07	1997-08-04	1997-07-14	3	44.77	Mère Paillarde	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada
10591	VAFFE	1	1997-07-07	1997-07-21	1997-07-16	1	55.92	Vaffeljernet	Smagsloget 45	Århus	\N	8200	Denmark
10592	LEHMS	3	1997-07-08	1997-08-05	1997-07-16	1	32.1	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10593	LEHMS	7	1997-07-09	1997-08-06	1997-08-13	2	174.2	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10594	OLDWO	3	1997-07-09	1997-08-06	1997-07-16	2	5.24	Old World Delicatessen	2743 Bering St.	Anchorage	AK	99508	USA
10595	ERNSH	2	1997-07-10	1997-08-07	1997-07-14	1	96.78	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10596	WHITC	8	1997-07-11	1997-08-08	1997-08-12	1	16.34	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
10597	PICCO	7	1997-07-11	1997-08-08	1997-07-18	3	35.12	Piccolo und mehr	Geislweg 14	Salzburg	\N	5020	Austria
10598	RATTC	1	1997-07-14	1997-08-11	1997-07-18	3	44.42	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10599	BSBEV	6	1997-07-15	1997-08-26	1997-07-21	3	29.98	B's Beverages	Fauntleroy Circus	London	\N	EC2 5NT	UK
10600	HUNGC	4	1997-07-16	1997-08-13	1997-07-21	1	45.13	Hungry Coyote Import Store	City Center Plaza 516 Main St.	Elgin	OR	97827	USA
10601	HILAA	7	1997-07-16	1997-08-27	1997-07-22	1	58.3	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10602	VAFFE	8	1997-07-17	1997-08-14	1997-07-22	2	2.92	Vaffeljernet	Smagsloget 45	Århus	\N	8200	Denmark
10603	SAVEA	8	1997-07-18	1997-08-15	1997-08-08	2	48.77	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10604	FURIB	1	1997-07-18	1997-08-15	1997-07-29	1	7.46	Furia Bacalhau e Frutos do Mar	Jardim das rosas n. 32	Lisboa	\N	1675	Portugal
10605	MEREP	1	1997-07-21	1997-08-18	1997-07-29	2	379.13	Mère Paillarde	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada
10606	TRADH	4	1997-07-22	1997-08-19	1997-07-31	3	79.4	Tradiçao Hipermercados	Av. Inês de Castro, 414	Sao Paulo	SP	05634-030	Brazil
10607	SAVEA	5	1997-07-22	1997-08-19	1997-07-25	1	200.24	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10608	TOMSP	4	1997-07-23	1997-08-20	1997-08-01	2	27.79	Toms Spezialitäten	Luisenstr. 48	Münster	\N	44087	Germany
10609	DUMON	7	1997-07-24	1997-08-21	1997-07-30	2	1.85	Du monde entier	67, rue des Cinquante Otages	Nantes	\N	44000	France
10610	LAMAI	8	1997-07-25	1997-08-22	1997-08-06	1	26.78	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
10611	WOLZA	6	1997-07-25	1997-08-22	1997-08-01	2	80.65	Wolski Zajazd	ul. Filtrowa 68	Warszawa	\N	01-012	Poland
10612	SAVEA	1	1997-07-28	1997-08-25	1997-08-01	2	544.08	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10613	HILAA	4	1997-07-29	1997-08-26	1997-08-01	2	8.11	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10614	BLAUS	8	1997-07-29	1997-08-26	1997-08-01	3	1.93	Blauer See Delikatessen	Forsterstr. 57	Mannheim	\N	68306	Germany
10615	WILMK	2	1997-07-30	1997-08-27	1997-08-06	3	0.75	Wilman Kala	Keskuskatu 45	Helsinki	\N	21240	Finland
10616	GREAL	1	1997-07-31	1997-08-28	1997-08-05	2	116.53	Great Lakes Food Market	2732 Baker Blvd.	Eugene	OR	97403	USA
10617	GREAL	4	1997-07-31	1997-08-28	1997-08-04	2	18.53	Great Lakes Food Market	2732 Baker Blvd.	Eugene	OR	97403	USA
10618	MEREP	1	1997-08-01	1997-09-12	1997-08-08	1	154.68	Mère Paillarde	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada
10619	MEREP	3	1997-08-04	1997-09-01	1997-08-07	3	91.05	Mère Paillarde	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada
10620	LAUGB	2	1997-08-05	1997-09-02	1997-08-14	3	0.94	Laughing Bacchus Wine Cellars	2319 Elm St.	Vancouver	BC	V3F 2K1	Canada
10621	ISLAT	4	1997-08-05	1997-09-02	1997-08-11	2	23.73	Island Trading	Garden House Crowther Way	Cowes	Isle of Wight	PO31 7PJ	UK
10622	RICAR	4	1997-08-06	1997-09-03	1997-08-11	3	50.97	Ricardo Adocicados	Av. Copacabana, 267	Rio de Janeiro	RJ	02389-890	Brazil
10623	FRANK	8	1997-08-07	1997-09-04	1997-08-12	2	97.18	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10624	THECR	4	1997-08-07	1997-09-04	1997-08-19	2	94.8	The Cracker Box	55 Grizzly Peak Rd.	Butte	MT	59801	USA
10625	ANATR	3	1997-08-08	1997-09-05	1997-08-14	1	43.9	Ana Trujillo Emparedados y helados	Avda. de la Constitución 2222	México D.F.	\N	05021	Mexico
10626	BERGS	1	1997-08-11	1997-09-08	1997-08-20	2	138.69	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10627	SAVEA	8	1997-08-11	1997-09-22	1997-08-21	3	107.46	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10628	BLONP	4	1997-08-12	1997-09-09	1997-08-20	3	30.36	Blondel père et fils	24, place Kléber	Strasbourg	\N	67000	France
10629	GODOS	4	1997-08-12	1997-09-09	1997-08-20	3	85.46	Godos Cocina Típica	C/ Romero, 33	Sevilla	\N	41101	Spain
10630	KOENE	1	1997-08-13	1997-09-10	1997-08-19	2	32.35	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
10631	LAMAI	8	1997-08-14	1997-09-11	1997-08-15	1	0.87	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
10632	WANDK	8	1997-08-14	1997-09-11	1997-08-19	1	41.38	Die Wandernde Kuh	Adenauerallee 900	Stuttgart	\N	70563	Germany
10633	ERNSH	7	1997-08-15	1997-09-12	1997-08-18	3	477.9	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10634	FOLIG	4	1997-08-15	1997-09-12	1997-08-21	3	487.38	Folies gourmandes	184, chaussée de Tournai	Lille	\N	59000	France
10635	MAGAA	8	1997-08-18	1997-09-15	1997-08-21	3	47.46	Magazzini Alimentari Riuniti	Via Ludovico il Moro 22	Bergamo	\N	24100	Italy
10636	WARTH	4	1997-08-19	1997-09-16	1997-08-26	1	1.15	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10637	QUEEN	6	1997-08-19	1997-09-16	1997-08-26	1	201.29	Queen Cozinha	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil
10638	LINOD	3	1997-08-20	1997-09-17	1997-09-01	1	158.44	LINO-Delicateses	Ave. 5 de Mayo Porlamar	I. de Margarita	Nueva Esparta	4980	Venezuela
10639	SANTG	7	1997-08-20	1997-09-17	1997-08-27	3	38.64	Santé Gourmet	Erling Skakkes gate 78	Stavern	\N	4110	Norway
10640	WANDK	4	1997-08-21	1997-09-18	1997-08-28	1	23.55	Die Wandernde Kuh	Adenauerallee 900	Stuttgart	\N	70563	Germany
10641	HILAA	4	1997-08-22	1997-09-19	1997-08-26	2	179.61	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10642	SIMOB	7	1997-08-22	1997-09-19	1997-09-05	3	41.89	Simons bistro	Vinbæltet 34	Kobenhavn	\N	1734	Denmark
10643	ALFKI	6	1997-08-25	1997-09-22	1997-09-02	1	29.46	Alfreds Futterkiste	Obere Str. 57	Berlin	\N	12209	Germany
10644	WELLI	3	1997-08-25	1997-09-22	1997-09-01	2	0.14	Wellington Importadora	Rua do Mercado, 12	Resende	SP	08737-363	Brazil
10645	HANAR	4	1997-08-26	1997-09-23	1997-09-02	1	12.41	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
10646	HUNGO	9	1997-08-27	1997-10-08	1997-09-03	3	142.33	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10647	QUEDE	4	1997-08-27	1997-09-10	1997-09-03	2	45.54	Que Delícia	Rua da Panificadora, 12	Rio de Janeiro	RJ	02389-673	Brazil
10648	RICAR	5	1997-08-28	1997-10-09	1997-09-09	2	14.25	Ricardo Adocicados	Av. Copacabana, 267	Rio de Janeiro	RJ	02389-890	Brazil
10649	MAISD	5	1997-08-28	1997-09-25	1997-08-29	3	6.2	Maison Dewey	Rue Joseph-Bens 532	Bruxelles	\N	B-1180	Belgium
10650	FAMIA	5	1997-08-29	1997-09-26	1997-09-03	3	176.81	Familia Arquibaldo	Rua Orós, 92	Sao Paulo	SP	05442-030	Brazil
10651	WANDK	8	1997-09-01	1997-09-29	1997-09-11	2	20.6	Die Wandernde Kuh	Adenauerallee 900	Stuttgart	\N	70563	Germany
10652	GOURL	4	1997-09-01	1997-09-29	1997-09-08	2	7.14	Gourmet Lanchonetes	Av. Brasil, 442	Campinas	SP	04876-786	Brazil
10653	FRANK	1	1997-09-02	1997-09-30	1997-09-19	1	93.25	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10654	BERGS	5	1997-09-02	1997-09-30	1997-09-11	1	55.26	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10655	REGGC	1	1997-09-03	1997-10-01	1997-09-11	2	4.41	Reggiani Caseifici	Strada Provinciale 124	Reggio Emilia	\N	42100	Italy
10656	GREAL	6	1997-09-04	1997-10-02	1997-09-10	1	57.15	Great Lakes Food Market	2732 Baker Blvd.	Eugene	OR	97403	USA
10657	SAVEA	2	1997-09-04	1997-10-02	1997-09-15	2	352.69	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10658	QUICK	4	1997-09-05	1997-10-03	1997-09-08	1	364.15	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10659	QUEEN	7	1997-09-05	1997-10-03	1997-09-10	2	105.81	Queen Cozinha	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil
10660	HUNGC	8	1997-09-08	1997-10-06	1997-10-15	1	111.29	Hungry Coyote Import Store	City Center Plaza 516 Main St.	Elgin	OR	97827	USA
10661	HUNGO	7	1997-09-09	1997-10-07	1997-09-15	3	17.55	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10662	LONEP	3	1997-09-09	1997-10-07	1997-09-18	2	1.28	Lonesome Pine Restaurant	89 Chiaroscuro Rd.	Portland	OR	97219	USA
10663	BONAP	2	1997-09-10	1997-09-24	1997-10-03	2	113.15	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10664	FURIB	1	1997-09-10	1997-10-08	1997-09-19	3	1.27	Furia Bacalhau e Frutos do Mar	Jardim das rosas n. 32	Lisboa	\N	1675	Portugal
10665	LONEP	1	1997-09-11	1997-10-09	1997-09-17	2	26.31	Lonesome Pine Restaurant	89 Chiaroscuro Rd.	Portland	OR	97219	USA
10666	RICSU	7	1997-09-12	1997-10-10	1997-09-22	2	232.42	Richter Supermarkt	Starenweg 5	Genève	\N	1204	Switzerland
10667	ERNSH	7	1997-09-12	1997-10-10	1997-09-19	1	78.09	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10668	WANDK	1	1997-09-15	1997-10-13	1997-09-23	2	47.22	Die Wandernde Kuh	Adenauerallee 900	Stuttgart	\N	70563	Germany
10669	SIMOB	2	1997-09-15	1997-10-13	1997-09-22	1	24.39	Simons bistro	Vinbæltet 34	Kobenhavn	\N	1734	Denmark
10670	FRANK	4	1997-09-16	1997-10-14	1997-09-18	1	203.48	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10671	FRANR	1	1997-09-17	1997-10-15	1997-09-24	1	30.34	France restauration	54, rue Royale	Nantes	\N	44000	France
10672	BERGS	9	1997-09-17	1997-10-01	1997-09-26	2	95.75	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10673	WILMK	2	1997-09-18	1997-10-16	1997-09-19	1	22.76	Wilman Kala	Keskuskatu 45	Helsinki	\N	21240	Finland
10674	ISLAT	4	1997-09-18	1997-10-16	1997-09-30	2	0.9	Island Trading	Garden House Crowther Way	Cowes	Isle of Wight	PO31 7PJ	UK
10675	FRANK	5	1997-09-19	1997-10-17	1997-09-23	2	31.85	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10676	TORTU	2	1997-09-22	1997-10-20	1997-09-29	2	2.01	Tortuga Restaurante	Avda. Azteca 123	México D.F.	\N	05033	Mexico
10677	ANTON	1	1997-09-22	1997-10-20	1997-09-26	3	4.03	Antonio Moreno Taquería	Mataderos  2312	México D.F.	\N	05023	Mexico
10678	SAVEA	7	1997-09-23	1997-10-21	1997-10-16	3	388.98	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10679	BLONP	8	1997-09-23	1997-10-21	1997-09-30	3	27.94	Blondel père et fils	24, place Kléber	Strasbourg	\N	67000	France
10680	OLDWO	1	1997-09-24	1997-10-22	1997-09-26	1	26.61	Old World Delicatessen	2743 Bering St.	Anchorage	AK	99508	USA
10681	GREAL	3	1997-09-25	1997-10-23	1997-09-30	3	76.13	Great Lakes Food Market	2732 Baker Blvd.	Eugene	OR	97403	USA
10682	ANTON	3	1997-09-25	1997-10-23	1997-10-01	2	36.13	Antonio Moreno Taquería	Mataderos  2312	México D.F.	\N	05023	Mexico
10683	DUMON	2	1997-09-26	1997-10-24	1997-10-01	1	4.4	Du monde entier	67, rue des Cinquante Otages	Nantes	\N	44000	France
10684	OTTIK	3	1997-09-26	1997-10-24	1997-09-30	1	145.63	Ottilies Käseladen	Mehrheimerstr. 369	Köln	\N	50739	Germany
10685	GOURL	4	1997-09-29	1997-10-13	1997-10-03	2	33.75	Gourmet Lanchonetes	Av. Brasil, 442	Campinas	SP	04876-786	Brazil
10686	PICCO	2	1997-09-30	1997-10-28	1997-10-08	1	96.5	Piccolo und mehr	Geislweg 14	Salzburg	\N	5020	Austria
10687	HUNGO	9	1997-09-30	1997-10-28	1997-10-30	2	296.43	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10688	VAFFE	4	1997-10-01	1997-10-15	1997-10-07	2	299.09	Vaffeljernet	Smagsloget 45	Århus	\N	8200	Denmark
10689	BERGS	1	1997-10-01	1997-10-29	1997-10-07	2	13.42	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10690	HANAR	1	1997-10-02	1997-10-30	1997-10-03	1	15.8	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
10691	QUICK	2	1997-10-03	1997-11-14	1997-10-22	2	810.05	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10692	ALFKI	4	1997-10-03	1997-10-31	1997-10-13	2	61.02	Alfred's Futterkiste	Obere Str. 57	Berlin	\N	12209	Germany
10693	WHITC	3	1997-10-06	1997-10-20	1997-10-10	3	139.34	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
10694	QUICK	8	1997-10-06	1997-11-03	1997-10-09	3	398.36	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10695	WILMK	7	1997-10-07	1997-11-18	1997-10-14	1	16.72	Wilman Kala	Keskuskatu 45	Helsinki	\N	21240	Finland
10696	WHITC	8	1997-10-08	1997-11-19	1997-10-14	3	102.55	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
10697	LINOD	3	1997-10-08	1997-11-05	1997-10-14	1	45.52	LINO-Delicateses	Ave. 5 de Mayo Porlamar	I. de Margarita	Nueva Esparta	4980	Venezuela
10698	ERNSH	4	1997-10-09	1997-11-06	1997-10-17	1	272.47	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10699	MORGK	3	1997-10-09	1997-11-06	1997-10-13	3	0.58	Morgenstern Gesundkost	Heerstr. 22	Leipzig	\N	04179	Germany
10700	SAVEA	3	1997-10-10	1997-11-07	1997-10-16	1	65.1	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10701	HUNGO	6	1997-10-13	1997-10-27	1997-10-15	3	220.31	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10702	ALFKI	4	1997-10-13	1997-11-24	1997-10-21	1	23.94	Alfred's Futterkiste	Obere Str. 57	Berlin	\N	12209	Germany
10703	FOLKO	6	1997-10-14	1997-11-11	1997-10-20	2	152.3	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10704	QUEEN	6	1997-10-14	1997-11-11	1997-11-07	1	4.78	Queen Cozinha	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil
10705	HILAA	9	1997-10-15	1997-11-12	1997-11-18	2	3.52	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10706	OLDWO	8	1997-10-16	1997-11-13	1997-10-21	3	135.63	Old World Delicatessen	2743 Bering St.	Anchorage	AK	99508	USA
10707	AROUT	4	1997-10-16	1997-10-30	1997-10-23	3	21.74	Around the Horn	Brook Farm Stratford St. Mary	Colchester	Essex	CO7 6JX	UK
10708	THEBI	6	1997-10-17	1997-11-28	1997-11-05	2	2.96	The Big Cheese	89 Jefferson Way Suite 2	Portland	OR	97201	USA
10709	GOURL	1	1997-10-17	1997-11-14	1997-11-20	3	210.8	Gourmet Lanchonetes	Av. Brasil, 442	Campinas	SP	04876-786	Brazil
10710	FRANS	1	1997-10-20	1997-11-17	1997-10-23	1	4.98	Franchi S.p.A.	Via Monte Bianco 34	Torino	\N	10100	Italy
10711	SAVEA	5	1997-10-21	1997-12-02	1997-10-29	2	52.41	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10712	HUNGO	3	1997-10-21	1997-11-18	1997-10-31	1	89.93	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10713	SAVEA	1	1997-10-22	1997-11-19	1997-10-24	1	167.05	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10714	SAVEA	5	1997-10-22	1997-11-19	1997-10-27	3	24.49	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10715	BONAP	3	1997-10-23	1997-11-06	1997-10-29	1	63.2	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10716	RANCH	4	1997-10-24	1997-11-21	1997-10-27	2	22.57	Rancho grande	Av. del Libertador 900	Buenos Aires	\N	1010	Argentina
10717	FRANK	1	1997-10-24	1997-11-21	1997-10-29	2	59.25	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10718	KOENE	1	1997-10-27	1997-11-24	1997-10-29	3	170.88	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
10719	LETSS	8	1997-10-27	1997-11-24	1997-11-05	2	51.44	Let's Stop N Shop	87 Polk St. Suite 5	San Francisco	CA	94117	USA
10720	QUEDE	8	1997-10-28	1997-11-11	1997-11-05	2	9.53	Que Delícia	Rua da Panificadora, 12	Rio de Janeiro	RJ	02389-673	Brazil
10721	QUICK	5	1997-10-29	1997-11-26	1997-10-31	3	48.92	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10722	SAVEA	8	1997-10-29	1997-12-10	1997-11-04	1	74.58	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10723	WHITC	3	1997-10-30	1997-11-27	1997-11-25	1	21.72	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
10724	MEREP	8	1997-10-30	1997-12-11	1997-11-05	2	57.75	Mère Paillarde	43 rue St. Laurent	Montréal	Québec	H1J 1C3	Canada
10725	FAMIA	4	1997-10-31	1997-11-28	1997-11-05	3	10.83	Familia Arquibaldo	Rua Orós, 92	Sao Paulo	SP	05442-030	Brazil
10726	EASTC	4	1997-11-03	1997-11-17	1997-12-05	1	16.56	Eastern Connection	35 King George	London	\N	WX3 6FW	UK
10727	REGGC	2	1997-11-03	1997-12-01	1997-12-05	1	89.9	Reggiani Caseifici	Strada Provinciale 124	Reggio Emilia	\N	42100	Italy
10728	QUEEN	4	1997-11-04	1997-12-02	1997-11-11	2	58.33	Queen Cozinha	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil
10729	LINOD	8	1997-11-04	1997-12-16	1997-11-14	3	141.06	LINO-Delicateses	Ave. 5 de Mayo Porlamar	I. de Margarita	Nueva Esparta	4980	Venezuela
10730	BONAP	5	1997-11-05	1997-12-03	1997-11-14	1	20.12	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10731	CHOPS	7	1997-11-06	1997-12-04	1997-11-14	1	96.65	Chop-suey Chinese	Hauptstr. 31	Bern	\N	3012	Switzerland
10732	BONAP	3	1997-11-06	1997-12-04	1997-11-07	1	16.97	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10733	BERGS	1	1997-11-07	1997-12-05	1997-11-10	3	110.11	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10734	GOURL	2	1997-11-07	1997-12-05	1997-11-12	3	1.63	Gourmet Lanchonetes	Av. Brasil, 442	Campinas	SP	04876-786	Brazil
10735	LETSS	6	1997-11-10	1997-12-08	1997-11-21	2	45.97	Let's Stop N Shop	87 Polk St. Suite 5	San Francisco	CA	94117	USA
10736	HUNGO	9	1997-11-11	1997-12-09	1997-11-21	2	44.1	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10737	VINET	2	1997-11-11	1997-12-09	1997-11-18	2	7.79	Vins et alcools Chevalier	59 rue de l'Abbaye	Reims	\N	51100	France
10738	SPECD	2	1997-11-12	1997-12-10	1997-11-18	1	2.91	Spécialités du monde	25, rue Lauriston	Paris	\N	75016	France
10739	VINET	3	1997-11-12	1997-12-10	1997-11-17	3	11.08	Vins et alcools Chevalier	59 rue de l'Abbaye	Reims	\N	51100	France
10740	WHITC	4	1997-11-13	1997-12-11	1997-11-25	2	81.88	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
10741	AROUT	4	1997-11-14	1997-11-28	1997-11-18	3	10.96	Around the Horn	Brook Farm Stratford St. Mary	Colchester	Essex	CO7 6JX	UK
10742	BOTTM	3	1997-11-14	1997-12-12	1997-11-18	3	243.73	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
10743	AROUT	1	1997-11-17	1997-12-15	1997-11-21	2	23.72	Around the Horn	Brook Farm Stratford St. Mary	Colchester	Essex	CO7 6JX	UK
10744	VAFFE	6	1997-11-17	1997-12-15	1997-11-24	1	69.19	Vaffeljernet	Smagsloget 45	Århus	\N	8200	Denmark
10745	QUICK	9	1997-11-18	1997-12-16	1997-11-27	1	3.52	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10746	CHOPS	1	1997-11-19	1997-12-17	1997-11-21	3	31.43	Chop-suey Chinese	Hauptstr. 31	Bern	\N	3012	Switzerland
10747	PICCO	6	1997-11-19	1997-12-17	1997-11-26	1	117.33	Piccolo und mehr	Geislweg 14	Salzburg	\N	5020	Austria
10748	SAVEA	3	1997-11-20	1997-12-18	1997-11-28	1	232.55	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10749	ISLAT	4	1997-11-20	1997-12-18	1997-12-19	2	61.53	Island Trading	Garden House Crowther Way	Cowes	Isle of Wight	PO31 7PJ	UK
10750	WARTH	9	1997-11-21	1997-12-19	1997-11-24	1	79.3	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10751	RICSU	3	1997-11-24	1997-12-22	1997-12-03	3	130.79	Richter Supermarkt	Starenweg 5	Genève	\N	1204	Switzerland
10752	NORTS	2	1997-11-24	1997-12-22	1997-11-28	3	1.39	North/South	South House 300 Queensbridge	London	\N	SW7 1RZ	UK
10753	FRANS	3	1997-11-25	1997-12-23	1997-11-27	1	7.7	Franchi S.p.A.	Via Monte Bianco 34	Torino	\N	10100	Italy
10754	MAGAA	6	1997-11-25	1997-12-23	1997-11-27	3	2.38	Magazzini Alimentari Riuniti	Via Ludovico il Moro 22	Bergamo	\N	24100	Italy
10755	BONAP	4	1997-11-26	1997-12-24	1997-11-28	2	16.71	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10756	SPLIR	8	1997-11-27	1997-12-25	1997-12-02	2	73.21	Split Rail Beer & Ale	P.O. Box 555	Lander	WY	82520	USA
10757	SAVEA	6	1997-11-27	1997-12-25	1997-12-15	1	8.19	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10758	RICSU	3	1997-11-28	1997-12-26	1997-12-04	3	138.17	Richter Supermarkt	Starenweg 5	Genève	\N	1204	Switzerland
10759	ANATR	3	1997-11-28	1997-12-26	1997-12-12	3	11.99	Ana Trujillo Emparedados y helados	Avda. de la Constitución 2222	México D.F.	\N	05021	Mexico
10760	MAISD	4	1997-12-01	1997-12-29	1997-12-10	1	155.64	Maison Dewey	Rue Joseph-Bens 532	Bruxelles	\N	B-1180	Belgium
10761	RATTC	5	1997-12-02	1997-12-30	1997-12-08	2	18.66	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10762	FOLKO	3	1997-12-02	1997-12-30	1997-12-09	1	328.74	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10763	FOLIG	3	1997-12-03	1997-12-31	1997-12-08	3	37.35	Folies gourmandes	184, chaussée de Tournai	Lille	\N	59000	France
10764	ERNSH	6	1997-12-03	1997-12-31	1997-12-08	3	145.45	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10765	QUICK	3	1997-12-04	1998-01-01	1997-12-09	3	42.74	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10766	OTTIK	4	1997-12-05	1998-01-02	1997-12-09	1	157.55	Ottilies Käseladen	Mehrheimerstr. 369	Köln	\N	50739	Germany
10767	SUPRD	4	1997-12-05	1998-01-02	1997-12-15	3	1.59	Suprêmes délices	Boulevard Tirou, 255	Charleroi	\N	B-6000	Belgium
10768	AROUT	3	1997-12-08	1998-01-05	1997-12-15	2	146.32	Around the Horn	Brook Farm Stratford St. Mary	Colchester	Essex	CO7 6JX	UK
10769	VAFFE	3	1997-12-08	1998-01-05	1997-12-12	1	65.06	Vaffeljernet	Smagsloget 45	Århus	\N	8200	Denmark
10770	HANAR	8	1997-12-09	1998-01-06	1997-12-17	3	5.32	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
10771	ERNSH	9	1997-12-10	1998-01-07	1998-01-02	2	11.19	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10772	LEHMS	3	1997-12-10	1998-01-07	1997-12-19	2	91.28	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10773	ERNSH	1	1997-12-11	1998-01-08	1997-12-16	3	96.43	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10774	FOLKO	4	1997-12-11	1997-12-25	1997-12-12	1	48.2	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10775	THECR	7	1997-12-12	1998-01-09	1997-12-26	1	20.25	The Cracker Box	55 Grizzly Peak Rd.	Butte	MT	59801	USA
10776	ERNSH	1	1997-12-15	1998-01-12	1997-12-18	3	351.53	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10777	GOURL	7	1997-12-15	1997-12-29	1998-01-21	2	3.01	Gourmet Lanchonetes	Av. Brasil, 442	Campinas	SP	04876-786	Brazil
10778	BERGS	3	1997-12-16	1998-01-13	1997-12-24	1	6.79	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10779	MORGK	3	1997-12-16	1998-01-13	1998-01-14	2	58.13	Morgenstern Gesundkost	Heerstr. 22	Leipzig	\N	04179	Germany
10780	LILAS	2	1997-12-16	1997-12-30	1997-12-25	1	42.13	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
10781	WARTH	2	1997-12-17	1998-01-14	1997-12-19	3	73.16	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
10782	CACTU	9	1997-12-17	1998-01-14	1997-12-22	3	1.1	Cactus Comidas para llevar	Cerrito 333	Buenos Aires	\N	1010	Argentina
10783	HANAR	4	1997-12-18	1998-01-15	1997-12-19	2	124.98	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
10784	MAGAA	4	1997-12-18	1998-01-15	1997-12-22	3	70.09	Magazzini Alimentari Riuniti	Via Ludovico il Moro 22	Bergamo	\N	24100	Italy
10785	GROSR	1	1997-12-18	1998-01-15	1997-12-24	3	1.51	GROSELLA-Restaurante	5ª Ave. Los Palos Grandes	Caracas	DF	1081	Venezuela
10786	QUEEN	8	1997-12-19	1998-01-16	1997-12-23	1	110.87	Queen Cozinha	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil
10787	LAMAI	2	1997-12-19	1998-01-02	1997-12-26	1	249.93	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
10788	QUICK	1	1997-12-22	1998-01-19	1998-01-19	2	42.7	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10789	FOLIG	1	1997-12-22	1998-01-19	1997-12-31	2	100.6	Folies gourmandes	184, chaussée de Tournai	Lille	\N	59000	France
10790	GOURL	6	1997-12-22	1998-01-19	1997-12-26	1	28.23	Gourmet Lanchonetes	Av. Brasil, 442	Campinas	SP	04876-786	Brazil
10791	FRANK	6	1997-12-23	1998-01-20	1998-01-01	2	16.85	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10792	WOLZA	1	1997-12-23	1998-01-20	1997-12-31	3	23.79	Wolski Zajazd	ul. Filtrowa 68	Warszawa	\N	01-012	Poland
10793	AROUT	3	1997-12-24	1998-01-21	1998-01-08	3	4.52	Around the Horn	Brook Farm Stratford St. Mary	Colchester	Essex	CO7 6JX	UK
10794	QUEDE	6	1997-12-24	1998-01-21	1998-01-02	1	21.49	Que Delícia	Rua da Panificadora, 12	Rio de Janeiro	RJ	02389-673	Brazil
10795	ERNSH	8	1997-12-24	1998-01-21	1998-01-20	2	126.66	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10796	HILAA	3	1997-12-25	1998-01-22	1998-01-14	1	26.52	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10797	DRACD	7	1997-12-25	1998-01-22	1998-01-05	2	33.35	Drachenblut Delikatessen	Walserweg 21	Aachen	\N	52066	Germany
10798	ISLAT	2	1997-12-26	1998-01-23	1998-01-05	1	2.33	Island Trading	Garden House Crowther Way	Cowes	Isle of Wight	PO31 7PJ	UK
10799	KOENE	9	1997-12-26	1998-02-06	1998-01-05	3	30.76	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
10800	SEVES	1	1997-12-26	1998-01-23	1998-01-05	3	137.44	Seven Seas Imports	90 Wadhurst Rd.	London	\N	OX15 4NB	UK
10801	BOLID	4	1997-12-29	1998-01-26	1997-12-31	2	97.09	Bólido Comidas preparadas	C/ Araquil, 67	Madrid	\N	28023	Spain
10802	SIMOB	4	1997-12-29	1998-01-26	1998-01-02	2	257.26	Simons bistro	Vinbæltet 34	Kobenhavn	\N	1734	Denmark
10803	WELLI	4	1997-12-30	1998-01-27	1998-01-06	1	55.23	Wellington Importadora	Rua do Mercado, 12	Resende	SP	08737-363	Brazil
10804	SEVES	6	1997-12-30	1998-01-27	1998-01-07	2	27.33	Seven Seas Imports	90 Wadhurst Rd.	London	\N	OX15 4NB	UK
10805	THEBI	2	1997-12-30	1998-01-27	1998-01-09	3	237.34	The Big Cheese	89 Jefferson Way Suite 2	Portland	OR	97201	USA
10806	VICTE	3	1997-12-31	1998-01-28	1998-01-05	2	22.11	Victuailles en stock	2, rue du Commerce	Lyon	\N	69004	France
10807	FRANS	4	1997-12-31	1998-01-28	1998-01-30	1	1.36	Franchi S.p.A.	Via Monte Bianco 34	Torino	\N	10100	Italy
10808	OLDWO	2	1998-01-01	1998-01-29	1998-01-09	3	45.53	Old World Delicatessen	2743 Bering St.	Anchorage	AK	99508	USA
10809	WELLI	7	1998-01-01	1998-01-29	1998-01-07	1	4.87	Wellington Importadora	Rua do Mercado, 12	Resende	SP	08737-363	Brazil
10810	LAUGB	2	1998-01-01	1998-01-29	1998-01-07	3	4.33	Laughing Bacchus Wine Cellars	2319 Elm St.	Vancouver	BC	V3F 2K1	Canada
10811	LINOD	8	1998-01-02	1998-01-30	1998-01-08	1	31.22	LINO-Delicateses	Ave. 5 de Mayo Porlamar	I. de Margarita	Nueva Esparta	4980	Venezuela
10812	REGGC	5	1998-01-02	1998-01-30	1998-01-12	1	59.78	Reggiani Caseifici	Strada Provinciale 124	Reggio Emilia	\N	42100	Italy
10813	RICAR	1	1998-01-05	1998-02-02	1998-01-09	1	47.38	Ricardo Adocicados	Av. Copacabana, 267	Rio de Janeiro	RJ	02389-890	Brazil
10814	VICTE	3	1998-01-05	1998-02-02	1998-01-14	3	130.94	Victuailles en stock	2, rue du Commerce	Lyon	\N	69004	France
10815	SAVEA	2	1998-01-05	1998-02-02	1998-01-14	3	14.62	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10816	GREAL	4	1998-01-06	1998-02-03	1998-02-04	2	719.78	Great Lakes Food Market	2732 Baker Blvd.	Eugene	OR	97403	USA
10817	KOENE	3	1998-01-06	1998-01-20	1998-01-13	2	306.07	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
10818	MAGAA	7	1998-01-07	1998-02-04	1998-01-12	3	65.48	Magazzini Alimentari Riuniti	Via Ludovico il Moro 22	Bergamo	\N	24100	Italy
10819	CACTU	2	1998-01-07	1998-02-04	1998-01-16	3	19.76	Cactus Comidas para llevar	Cerrito 333	Buenos Aires	\N	1010	Argentina
10820	RATTC	3	1998-01-07	1998-02-04	1998-01-13	2	37.52	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10821	SPLIR	1	1998-01-08	1998-02-05	1998-01-15	1	36.68	Split Rail Beer & Ale	P.O. Box 555	Lander	WY	82520	USA
10822	TRAIH	6	1998-01-08	1998-02-05	1998-01-16	3	7	Trail's Head Gourmet Provisioners	722 DaVinci Blvd.	Kirkland	WA	98034	USA
10823	LILAS	5	1998-01-09	1998-02-06	1998-01-13	2	163.97	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
10824	FOLKO	8	1998-01-09	1998-02-06	1998-01-30	1	1.23	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10825	DRACD	1	1998-01-09	1998-02-06	1998-01-14	1	79.25	Drachenblut Delikatessen	Walserweg 21	Aachen	\N	52066	Germany
10826	BLONP	6	1998-01-12	1998-02-09	1998-02-06	1	7.09	Blondel père et fils	24, place Kléber	Strasbourg	\N	67000	France
10827	BONAP	1	1998-01-12	1998-01-26	1998-02-06	2	63.54	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10828	RANCH	9	1998-01-13	1998-01-27	1998-02-04	1	90.85	Rancho grande	Av. del Libertador 900	Buenos Aires	\N	1010	Argentina
10829	ISLAT	9	1998-01-13	1998-02-10	1998-01-23	1	154.72	Island Trading	Garden House Crowther Way	Cowes	Isle of Wight	PO31 7PJ	UK
10830	TRADH	4	1998-01-13	1998-02-24	1998-01-21	2	81.83	Tradiçao Hipermercados	Av. Inês de Castro, 414	Sao Paulo	SP	05634-030	Brazil
10831	SANTG	3	1998-01-14	1998-02-11	1998-01-23	2	72.19	Santé Gourmet	Erling Skakkes gate 78	Stavern	\N	4110	Norway
10832	LAMAI	2	1998-01-14	1998-02-11	1998-01-19	2	43.26	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
10833	OTTIK	6	1998-01-15	1998-02-12	1998-01-23	2	71.49	Ottilies Käseladen	Mehrheimerstr. 369	Köln	\N	50739	Germany
10834	TRADH	1	1998-01-15	1998-02-12	1998-01-19	3	29.78	Tradiçao Hipermercados	Av. Inês de Castro, 414	Sao Paulo	SP	05634-030	Brazil
10835	ALFKI	1	1998-01-15	1998-02-12	1998-01-21	3	69.53	Alfred's Futterkiste	Obere Str. 57	Berlin	\N	12209	Germany
10836	ERNSH	7	1998-01-16	1998-02-13	1998-01-21	1	411.88	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10837	BERGS	9	1998-01-16	1998-02-13	1998-01-23	3	13.32	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10838	LINOD	3	1998-01-19	1998-02-16	1998-01-23	3	59.28	LINO-Delicateses	Ave. 5 de Mayo Porlamar	I. de Margarita	Nueva Esparta	4980	Venezuela
10839	TRADH	3	1998-01-19	1998-02-16	1998-01-22	3	35.43	Tradiçao Hipermercados	Av. Inês de Castro, 414	Sao Paulo	SP	05634-030	Brazil
10840	LINOD	4	1998-01-19	1998-03-02	1998-02-16	2	2.71	LINO-Delicateses	Ave. 5 de Mayo Porlamar	I. de Margarita	Nueva Esparta	4980	Venezuela
10841	SUPRD	5	1998-01-20	1998-02-17	1998-01-29	2	424.3	Suprêmes délices	Boulevard Tirou, 255	Charleroi	\N	B-6000	Belgium
10842	TORTU	1	1998-01-20	1998-02-17	1998-01-29	3	54.42	Tortuga Restaurante	Avda. Azteca 123	México D.F.	\N	05033	Mexico
10843	VICTE	4	1998-01-21	1998-02-18	1998-01-26	2	9.26	Victuailles en stock	2, rue du Commerce	Lyon	\N	69004	France
10844	PICCO	8	1998-01-21	1998-02-18	1998-01-26	2	25.22	Piccolo und mehr	Geislweg 14	Salzburg	\N	5020	Austria
10845	QUICK	8	1998-01-21	1998-02-04	1998-01-30	1	212.98	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10846	SUPRD	2	1998-01-22	1998-03-05	1998-01-23	3	56.46	Suprêmes délices	Boulevard Tirou, 255	Charleroi	\N	B-6000	Belgium
10847	SAVEA	4	1998-01-22	1998-02-05	1998-02-10	3	487.57	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10848	CONSH	7	1998-01-23	1998-02-20	1998-01-29	2	38.24	Consolidated Holdings	Berkeley Gardens 12  Brewery	London	\N	WX1 6LT	UK
10849	KOENE	9	1998-01-23	1998-02-20	1998-01-30	2	0.56	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
10850	VICTE	1	1998-01-23	1998-03-06	1998-01-30	1	49.19	Victuailles en stock	2, rue du Commerce	Lyon	\N	69004	France
10851	RICAR	5	1998-01-26	1998-02-23	1998-02-02	1	160.55	Ricardo Adocicados	Av. Copacabana, 267	Rio de Janeiro	RJ	02389-890	Brazil
10852	RATTC	8	1998-01-26	1998-02-09	1998-01-30	1	174.05	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10853	BLAUS	9	1998-01-27	1998-02-24	1998-02-03	2	53.83	Blauer See Delikatessen	Forsterstr. 57	Mannheim	\N	68306	Germany
10854	ERNSH	3	1998-01-27	1998-02-24	1998-02-05	2	100.22	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10855	OLDWO	3	1998-01-27	1998-02-24	1998-02-04	1	170.97	Old World Delicatessen	2743 Bering St.	Anchorage	AK	99508	USA
10856	ANTON	3	1998-01-28	1998-02-25	1998-02-10	2	58.43	Antonio Moreno Taquería	Mataderos  2312	México D.F.	\N	05023	Mexico
10857	BERGS	8	1998-01-28	1998-02-25	1998-02-06	2	188.85	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10858	LACOR	2	1998-01-29	1998-02-26	1998-02-03	1	52.51	La corne d'abondance	67, avenue de l'Europe	Versailles	\N	78000	France
10859	FRANK	1	1998-01-29	1998-02-26	1998-02-02	2	76.1	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10860	FRANR	3	1998-01-29	1998-02-26	1998-02-04	3	19.26	France restauration	54, rue Royale	Nantes	\N	44000	France
10861	WHITC	4	1998-01-30	1998-02-27	1998-02-17	2	14.93	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
10862	LEHMS	8	1998-01-30	1998-03-13	1998-02-02	2	53.23	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10863	HILAA	4	1998-02-02	1998-03-02	1998-02-17	2	30.26	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10864	AROUT	4	1998-02-02	1998-03-02	1998-02-09	2	3.04	Around the Horn	Brook Farm Stratford St. Mary	Colchester	Essex	CO7 6JX	UK
10865	QUICK	2	1998-02-02	1998-02-16	1998-02-12	1	348.14	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10866	BERGS	5	1998-02-03	1998-03-03	1998-02-12	1	109.11	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10867	LONEP	6	1998-02-03	1998-03-17	1998-02-11	1	1.93	Lonesome Pine Restaurant	89 Chiaroscuro Rd.	Portland	OR	97219	USA
10868	QUEEN	7	1998-02-04	1998-03-04	1998-02-23	2	191.27	Queen Cozinha	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil
10869	SEVES	5	1998-02-04	1998-03-04	1998-02-09	1	143.28	Seven Seas Imports	90 Wadhurst Rd.	London	\N	OX15 4NB	UK
10870	WOLZA	5	1998-02-04	1998-03-04	1998-02-13	3	12.04	Wolski Zajazd	ul. Filtrowa 68	Warszawa	\N	01-012	Poland
10871	BONAP	9	1998-02-05	1998-03-05	1998-02-10	2	112.27	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10872	GODOS	5	1998-02-05	1998-03-05	1998-02-09	2	175.32	Godos Cocina Típica	C/ Romero, 33	Sevilla	\N	41101	Spain
10873	WILMK	4	1998-02-06	1998-03-06	1998-02-09	1	0.82	Wilman Kala	Keskuskatu 45	Helsinki	\N	21240	Finland
10874	GODOS	5	1998-02-06	1998-03-06	1998-02-11	2	19.58	Godos Cocina Típica	C/ Romero, 33	Sevilla	\N	41101	Spain
10875	BERGS	4	1998-02-06	1998-03-06	1998-03-03	2	32.37	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10876	BONAP	7	1998-02-09	1998-03-09	1998-02-12	3	60.42	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10877	RICAR	1	1998-02-09	1998-03-09	1998-02-19	1	38.06	Ricardo Adocicados	Av. Copacabana, 267	Rio de Janeiro	RJ	02389-890	Brazil
10878	QUICK	4	1998-02-10	1998-03-10	1998-02-12	1	46.69	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10879	WILMK	3	1998-02-10	1998-03-10	1998-02-12	3	8.5	Wilman Kala	Keskuskatu 45	Helsinki	\N	21240	Finland
10880	FOLKO	7	1998-02-10	1998-03-24	1998-02-18	1	88.01	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10881	CACTU	4	1998-02-11	1998-03-11	1998-02-18	1	2.84	Cactus Comidas para llevar	Cerrito 333	Buenos Aires	\N	1010	Argentina
10882	SAVEA	4	1998-02-11	1998-03-11	1998-02-20	3	23.1	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10883	LONEP	8	1998-02-12	1998-03-12	1998-02-20	3	0.53	Lonesome Pine Restaurant	89 Chiaroscuro Rd.	Portland	OR	97219	USA
10884	LETSS	4	1998-02-12	1998-03-12	1998-02-13	2	90.97	Let's Stop N Shop	87 Polk St. Suite 5	San Francisco	CA	94117	USA
10885	SUPRD	6	1998-02-12	1998-03-12	1998-02-18	3	5.64	Suprêmes délices	Boulevard Tirou, 255	Charleroi	\N	B-6000	Belgium
10886	HANAR	1	1998-02-13	1998-03-13	1998-03-02	1	4.99	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
10887	GALED	8	1998-02-13	1998-03-13	1998-02-16	3	1.25	Galería del gastronómo	Rambla de Cataluña, 23	Barcelona	\N	8022	Spain
10888	GODOS	1	1998-02-16	1998-03-16	1998-02-23	2	51.87	Godos Cocina Típica	C/ Romero, 33	Sevilla	\N	41101	Spain
10889	RATTC	9	1998-02-16	1998-03-16	1998-02-23	3	280.61	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10890	DUMON	7	1998-02-16	1998-03-16	1998-02-18	1	32.76	Du monde entier	67, rue des Cinquante Otages	Nantes	\N	44000	France
10891	LEHMS	7	1998-02-17	1998-03-17	1998-02-19	2	20.37	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10892	MAISD	4	1998-02-17	1998-03-17	1998-02-19	2	120.27	Maison Dewey	Rue Joseph-Bens 532	Bruxelles	\N	B-1180	Belgium
10893	KOENE	9	1998-02-18	1998-03-18	1998-02-20	2	77.78	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
10894	SAVEA	1	1998-02-18	1998-03-18	1998-02-20	1	116.13	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10895	ERNSH	3	1998-02-18	1998-03-18	1998-02-23	1	162.75	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10896	MAISD	7	1998-02-19	1998-03-19	1998-02-27	3	32.45	Maison Dewey	Rue Joseph-Bens 532	Bruxelles	\N	B-1180	Belgium
10897	HUNGO	3	1998-02-19	1998-03-19	1998-02-25	2	603.54	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10898	OCEAN	4	1998-02-20	1998-03-20	1998-03-06	2	1.27	Océano Atlántico Ltda.	Ing. Gustavo Moncada 8585 Piso 20-A	Buenos Aires	\N	1010	Argentina
10899	LILAS	5	1998-02-20	1998-03-20	1998-02-26	3	1.21	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
10900	WELLI	1	1998-02-20	1998-03-20	1998-03-04	2	1.66	Wellington Importadora	Rua do Mercado, 12	Resende	SP	08737-363	Brazil
10901	HILAA	4	1998-02-23	1998-03-23	1998-02-26	1	62.09	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10902	FOLKO	1	1998-02-23	1998-03-23	1998-03-03	1	44.15	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10903	HANAR	3	1998-02-24	1998-03-24	1998-03-04	3	36.71	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
10904	WHITC	3	1998-02-24	1998-03-24	1998-02-27	3	162.95	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
10905	WELLI	9	1998-02-24	1998-03-24	1998-03-06	2	13.72	Wellington Importadora	Rua do Mercado, 12	Resende	SP	08737-363	Brazil
10906	WOLZA	4	1998-02-25	1998-03-11	1998-03-03	3	26.29	Wolski Zajazd	ul. Filtrowa 68	Warszawa	\N	01-012	Poland
10907	SPECD	6	1998-02-25	1998-03-25	1998-02-27	3	9.19	Spécialités du monde	25, rue Lauriston	Paris	\N	75016	France
10908	REGGC	4	1998-02-26	1998-03-26	1998-03-06	2	32.96	Reggiani Caseifici	Strada Provinciale 124	Reggio Emilia	\N	42100	Italy
10909	SANTG	1	1998-02-26	1998-03-26	1998-03-10	2	53.05	Santé Gourmet	Erling Skakkes gate 78	Stavern	\N	4110	Norway
10910	WILMK	1	1998-02-26	1998-03-26	1998-03-04	3	38.11	Wilman Kala	Keskuskatu 45	Helsinki	\N	21240	Finland
10911	GODOS	3	1998-02-26	1998-03-26	1998-03-05	1	38.19	Godos Cocina Típica	C/ Romero, 33	Sevilla	\N	41101	Spain
10912	HUNGO	2	1998-02-26	1998-03-26	1998-03-18	2	580.91	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10913	QUEEN	4	1998-02-26	1998-03-26	1998-03-04	1	33.05	Queen Cozinha	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil
10914	QUEEN	6	1998-02-27	1998-03-27	1998-03-02	1	21.19	Queen Cozinha	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil
10915	TORTU	2	1998-02-27	1998-03-27	1998-03-02	2	3.51	Tortuga Restaurante	Avda. Azteca 123	México D.F.	\N	05033	Mexico
10916	RANCH	1	1998-02-27	1998-03-27	1998-03-09	2	63.77	Rancho grande	Av. del Libertador 900	Buenos Aires	\N	1010	Argentina
10917	ROMEY	4	1998-03-02	1998-03-30	1998-03-11	2	8.29	Romero y tomillo	Gran Vía, 1	Madrid	\N	28001	Spain
10918	BOTTM	3	1998-03-02	1998-03-30	1998-03-11	3	48.83	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
10919	LINOD	2	1998-03-02	1998-03-30	1998-03-04	2	19.8	LINO-Delicateses	Ave. 5 de Mayo Porlamar	I. de Margarita	Nueva Esparta	4980	Venezuela
10920	AROUT	4	1998-03-03	1998-03-31	1998-03-09	2	29.61	Around the Horn	Brook Farm Stratford St. Mary	Colchester	Essex	CO7 6JX	UK
10921	VAFFE	1	1998-03-03	1998-04-14	1998-03-09	1	176.48	Vaffeljernet	Smagsloget 45	Århus	\N	8200	Denmark
10922	HANAR	5	1998-03-03	1998-03-31	1998-03-05	3	62.74	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
10923	LAMAI	7	1998-03-03	1998-04-14	1998-03-13	3	68.26	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
10924	BERGS	3	1998-03-04	1998-04-01	1998-04-08	2	151.52	Berglunds snabbköp	Berguvsvägen  8	Luleå	\N	S-958 22	Sweden
10925	HANAR	3	1998-03-04	1998-04-01	1998-03-13	1	2.27	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
10926	ANATR	4	1998-03-04	1998-04-01	1998-03-11	3	39.92	Ana Trujillo Emparedados y helados	Avda. de la Constitución 2222	México D.F.	\N	05021	Mexico
10927	LACOR	4	1998-03-05	1998-04-02	1998-04-08	1	19.79	La corne d'abondance	67, avenue de l'Europe	Versailles	\N	78000	France
10928	GALED	1	1998-03-05	1998-04-02	1998-03-18	1	1.36	Galería del gastronómo	Rambla de Cataluña, 23	Barcelona	\N	8022	Spain
10929	FRANK	6	1998-03-05	1998-04-02	1998-03-12	1	33.93	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
10930	SUPRD	4	1998-03-06	1998-04-17	1998-03-18	3	15.55	Suprêmes délices	Boulevard Tirou, 255	Charleroi	\N	B-6000	Belgium
10931	RICSU	4	1998-03-06	1998-03-20	1998-03-19	2	13.6	Richter Supermarkt	Starenweg 5	Genève	\N	1204	Switzerland
10932	BONAP	8	1998-03-06	1998-04-03	1998-03-24	1	134.64	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10933	ISLAT	6	1998-03-06	1998-04-03	1998-03-16	3	54.15	Island Trading	Garden House Crowther Way	Cowes	Isle of Wight	PO31 7PJ	UK
10934	LEHMS	3	1998-03-09	1998-04-06	1998-03-12	3	32.01	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
10935	WELLI	4	1998-03-09	1998-04-06	1998-03-18	3	47.59	Wellington Importadora	Rua do Mercado, 12	Resende	SP	08737-363	Brazil
10936	GREAL	3	1998-03-09	1998-04-06	1998-03-18	2	33.68	Great Lakes Food Market	2732 Baker Blvd.	Eugene	OR	97403	USA
10937	CACTU	7	1998-03-10	1998-03-24	1998-03-13	3	31.51	Cactus Comidas para llevar	Cerrito 333	Buenos Aires	\N	1010	Argentina
10938	QUICK	3	1998-03-10	1998-04-07	1998-03-16	2	31.89	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10939	MAGAA	2	1998-03-10	1998-04-07	1998-03-13	2	76.33	Magazzini Alimentari Riuniti	Via Ludovico il Moro 22	Bergamo	\N	24100	Italy
10940	BONAP	8	1998-03-11	1998-04-08	1998-03-23	3	19.77	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
10941	SAVEA	7	1998-03-11	1998-04-08	1998-03-20	2	400.81	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10942	REGGC	9	1998-03-11	1998-04-08	1998-03-18	3	17.95	Reggiani Caseifici	Strada Provinciale 124	Reggio Emilia	\N	42100	Italy
10943	BSBEV	4	1998-03-11	1998-04-08	1998-03-19	2	2.17	B's Beverages	Fauntleroy Circus	London	\N	EC2 5NT	UK
10944	BOTTM	6	1998-03-12	1998-03-26	1998-03-13	3	52.92	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
10945	MORGK	4	1998-03-12	1998-04-09	1998-03-18	1	10.22	Morgenstern Gesundkost	Heerstr. 22	Leipzig	\N	04179	Germany
10946	VAFFE	1	1998-03-12	1998-04-09	1998-03-19	2	27.2	Vaffeljernet	Smagsloget 45	Århus	\N	8200	Denmark
10947	BSBEV	3	1998-03-13	1998-04-10	1998-03-16	2	3.26	B's Beverages	Fauntleroy Circus	London	\N	EC2 5NT	UK
10948	GODOS	3	1998-03-13	1998-04-10	1998-03-19	3	23.39	Godos Cocina Típica	C/ Romero, 33	Sevilla	\N	41101	Spain
10949	BOTTM	2	1998-03-13	1998-04-10	1998-03-17	3	74.44	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
10950	MAGAA	1	1998-03-16	1998-04-13	1998-03-23	2	2.5	Magazzini Alimentari Riuniti	Via Ludovico il Moro 22	Bergamo	\N	24100	Italy
10951	RICSU	9	1998-03-16	1998-04-27	1998-04-07	2	30.85	Richter Supermarkt	Starenweg 5	Genève	\N	1204	Switzerland
10952	ALFKI	1	1998-03-16	1998-04-27	1998-03-24	1	40.42	Alfred's Futterkiste	Obere Str. 57	Berlin	\N	12209	Germany
10953	AROUT	9	1998-03-16	1998-03-30	1998-03-25	2	23.72	Around the Horn	Brook Farm Stratford St. Mary	Colchester	Essex	CO7 6JX	UK
10954	LINOD	5	1998-03-17	1998-04-28	1998-03-20	1	27.91	LINO-Delicateses	Ave. 5 de Mayo Porlamar	I. de Margarita	Nueva Esparta	4980	Venezuela
10955	FOLKO	8	1998-03-17	1998-04-14	1998-03-20	2	3.26	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10956	BLAUS	6	1998-03-17	1998-04-28	1998-03-20	2	44.65	Blauer See Delikatessen	Forsterstr. 57	Mannheim	\N	68306	Germany
10957	HILAA	8	1998-03-18	1998-04-15	1998-03-27	3	105.36	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10958	OCEAN	7	1998-03-18	1998-04-15	1998-03-27	2	49.56	Océano Atlántico Ltda.	Ing. Gustavo Moncada 8585 Piso 20-A	Buenos Aires	\N	1010	Argentina
10959	GOURL	6	1998-03-18	1998-04-29	1998-03-23	2	4.98	Gourmet Lanchonetes	Av. Brasil, 442	Campinas	SP	04876-786	Brazil
10960	HILAA	3	1998-03-19	1998-04-02	1998-04-08	1	2.08	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10961	QUEEN	8	1998-03-19	1998-04-16	1998-03-30	1	104.47	Queen Cozinha	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil
10962	QUICK	8	1998-03-19	1998-04-16	1998-03-23	2	275.79	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10963	FURIB	9	1998-03-19	1998-04-16	1998-03-26	3	2.7	Furia Bacalhau e Frutos do Mar	Jardim das rosas n. 32	Lisboa	\N	1675	Portugal
10964	SPECD	3	1998-03-20	1998-04-17	1998-03-24	2	87.38	Spécialités du monde	25, rue Lauriston	Paris	\N	75016	France
10965	OLDWO	6	1998-03-20	1998-04-17	1998-03-30	3	144.38	Old World Delicatessen	2743 Bering St.	Anchorage	AK	99508	USA
10966	CHOPS	4	1998-03-20	1998-04-17	1998-04-08	1	27.19	Chop-suey Chinese	Hauptstr. 31	Bern	\N	3012	Switzerland
10967	TOMSP	2	1998-03-23	1998-04-20	1998-04-02	2	62.22	Toms Spezialitäten	Luisenstr. 48	Münster	\N	44087	Germany
10968	ERNSH	1	1998-03-23	1998-04-20	1998-04-01	3	74.6	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10969	COMMI	1	1998-03-23	1998-04-20	1998-03-30	2	0.21	Comércio Mineiro	Av. dos Lusíadas, 23	Sao Paulo	SP	05432-043	Brazil
10970	BOLID	9	1998-03-24	1998-04-07	1998-04-24	1	16.16	Bólido Comidas preparadas	C/ Araquil, 67	Madrid	\N	28023	Spain
10971	FRANR	2	1998-03-24	1998-04-21	1998-04-02	2	121.82	France restauration	54, rue Royale	Nantes	\N	44000	France
10972	LACOR	4	1998-03-24	1998-04-21	1998-03-26	2	0.02	La corne d'abondance	67, avenue de l'Europe	Versailles	\N	78000	France
10973	LACOR	6	1998-03-24	1998-04-21	1998-03-27	2	15.17	La corne d'abondance	67, avenue de l'Europe	Versailles	\N	78000	France
10974	SPLIR	3	1998-03-25	1998-04-08	1998-04-03	3	12.96	Split Rail Beer & Ale	P.O. Box 555	Lander	WY	82520	USA
10975	BOTTM	1	1998-03-25	1998-04-22	1998-03-27	3	32.27	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
10976	HILAA	1	1998-03-25	1998-05-06	1998-04-03	1	37.97	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
10977	FOLKO	8	1998-03-26	1998-04-23	1998-04-10	3	208.5	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10978	MAISD	9	1998-03-26	1998-04-23	1998-04-23	2	32.82	Maison Dewey	Rue Joseph-Bens 532	Bruxelles	\N	B-1180	Belgium
10979	ERNSH	8	1998-03-26	1998-04-23	1998-03-31	2	353.07	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10980	FOLKO	4	1998-03-27	1998-05-08	1998-04-17	1	1.26	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10981	HANAR	1	1998-03-27	1998-04-24	1998-04-02	2	193.37	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
10982	BOTTM	2	1998-03-27	1998-04-24	1998-04-08	1	14.01	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
10983	SAVEA	2	1998-03-27	1998-04-24	1998-04-06	2	657.54	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10984	SAVEA	1	1998-03-30	1998-04-27	1998-04-03	3	211.22	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
10985	HUNGO	2	1998-03-30	1998-04-27	1998-04-02	1	91.51	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
10986	OCEAN	8	1998-03-30	1998-04-27	1998-04-21	2	217.86	Océano Atlántico Ltda.	Ing. Gustavo Moncada 8585 Piso 20-A	Buenos Aires	\N	1010	Argentina
10987	EASTC	8	1998-03-31	1998-04-28	1998-04-06	1	185.48	Eastern Connection	35 King George	London	\N	WX3 6FW	UK
10988	RATTC	3	1998-03-31	1998-04-28	1998-04-10	2	61.14	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10989	QUEDE	2	1998-03-31	1998-04-28	1998-04-02	1	34.76	Que Delícia	Rua da Panificadora, 12	Rio de Janeiro	RJ	02389-673	Brazil
10990	ERNSH	2	1998-04-01	1998-05-13	1998-04-07	3	117.61	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
10991	QUICK	1	1998-04-01	1998-04-29	1998-04-07	1	38.51	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10992	THEBI	1	1998-04-01	1998-04-29	1998-04-03	3	4.27	The Big Cheese	89 Jefferson Way Suite 2	Portland	OR	97201	USA
10993	FOLKO	7	1998-04-01	1998-04-29	1998-04-10	3	8.81	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
10994	VAFFE	2	1998-04-02	1998-04-16	1998-04-09	3	65.53	Vaffeljernet	Smagsloget 45	Århus	\N	8200	Denmark
10995	PERIC	1	1998-04-02	1998-04-30	1998-04-06	3	46	Pericles Comidas clásicas	Calle Dr. Jorge Cash 321	México D.F.	\N	05033	Mexico
10996	QUICK	4	1998-04-02	1998-04-30	1998-04-10	2	1.12	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
10997	LILAS	8	1998-04-03	1998-05-15	1998-04-13	2	73.91	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
10998	WOLZA	8	1998-04-03	1998-04-17	1998-04-17	2	20.31	Wolski Zajazd	ul. Filtrowa 68	Warszawa	\N	01-012	Poland
10999	OTTIK	6	1998-04-03	1998-05-01	1998-04-10	2	96.35	Ottilies Käseladen	Mehrheimerstr. 369	Köln	\N	50739	Germany
11000	RATTC	2	1998-04-06	1998-05-04	1998-04-14	3	55.12	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
11001	FOLKO	2	1998-04-06	1998-05-04	1998-04-14	2	197.3	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
11002	SAVEA	4	1998-04-06	1998-05-04	1998-04-16	1	141.16	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
11003	THECR	3	1998-04-06	1998-05-04	1998-04-08	3	14.91	The Cracker Box	55 Grizzly Peak Rd.	Butte	MT	59801	USA
11004	MAISD	3	1998-04-07	1998-05-05	1998-04-20	1	44.84	Maison Dewey	Rue Joseph-Bens 532	Bruxelles	\N	B-1180	Belgium
11005	WILMK	2	1998-04-07	1998-05-05	1998-04-10	1	0.75	Wilman Kala	Keskuskatu 45	Helsinki	\N	21240	Finland
11006	GREAL	3	1998-04-07	1998-05-05	1998-04-15	2	25.19	Great Lakes Food Market	2732 Baker Blvd.	Eugene	OR	97403	USA
11007	PRINI	8	1998-04-08	1998-05-06	1998-04-13	2	202.24	Princesa Isabel Vinhos	Estrada da saúde n. 58	Lisboa	\N	1756	Portugal
11008	ERNSH	7	1998-04-08	1998-05-06	\N	3	79.46	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
11009	GODOS	2	1998-04-08	1998-05-06	1998-04-10	1	59.11	Godos Cocina Típica	C/ Romero, 33	Sevilla	\N	41101	Spain
11010	REGGC	2	1998-04-09	1998-05-07	1998-04-21	2	28.71	Reggiani Caseifici	Strada Provinciale 124	Reggio Emilia	\N	42100	Italy
11011	ALFKI	3	1998-04-09	1998-05-07	1998-04-13	1	1.21	Alfred's Futterkiste	Obere Str. 57	Berlin	\N	12209	Germany
11012	FRANK	1	1998-04-09	1998-04-23	1998-04-17	3	242.95	Frankenversand	Berliner Platz 43	München	\N	80805	Germany
11013	ROMEY	2	1998-04-09	1998-05-07	1998-04-10	1	32.99	Romero y tomillo	Gran Vía, 1	Madrid	\N	28001	Spain
11014	LINOD	2	1998-04-10	1998-05-08	1998-04-15	3	23.6	LINO-Delicateses	Ave. 5 de Mayo Porlamar	I. de Margarita	Nueva Esparta	4980	Venezuela
11015	SANTG	2	1998-04-10	1998-04-24	1998-04-20	2	4.62	Santé Gourmet	Erling Skakkes gate 78	Stavern	\N	4110	Norway
11016	AROUT	9	1998-04-10	1998-05-08	1998-04-13	2	33.8	Around the Horn	Brook Farm Stratford St. Mary	Colchester	Essex	CO7 6JX	UK
11017	ERNSH	9	1998-04-13	1998-05-11	1998-04-20	2	754.26	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
11018	LONEP	4	1998-04-13	1998-05-11	1998-04-16	2	11.65	Lonesome Pine Restaurant	89 Chiaroscuro Rd.	Portland	OR	97219	USA
11019	RANCH	6	1998-04-13	1998-05-11	\N	3	3.17	Rancho grande	Av. del Libertador 900	Buenos Aires	\N	1010	Argentina
11020	OTTIK	2	1998-04-14	1998-05-12	1998-04-16	2	43.3	Ottilies Käseladen	Mehrheimerstr. 369	Köln	\N	50739	Germany
11021	QUICK	3	1998-04-14	1998-05-12	1998-04-21	1	297.18	QUICK-Stop	Taucherstraße 10	Cunewalde	\N	01307	Germany
11022	HANAR	9	1998-04-14	1998-05-12	1998-05-04	2	6.27	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
11023	BSBEV	1	1998-04-14	1998-04-28	1998-04-24	2	123.83	B's Beverages	Fauntleroy Circus	London	\N	EC2 5NT	UK
11024	EASTC	4	1998-04-15	1998-05-13	1998-04-20	1	74.36	Eastern Connection	35 King George	London	\N	WX3 6FW	UK
11025	WARTH	6	1998-04-15	1998-05-13	1998-04-24	3	29.17	Wartian Herkku	Torikatu 38	Oulu	\N	90110	Finland
11026	FRANS	4	1998-04-15	1998-05-13	1998-04-28	1	47.09	Franchi S.p.A.	Via Monte Bianco 34	Torino	\N	10100	Italy
11027	BOTTM	1	1998-04-16	1998-05-14	1998-04-20	1	52.52	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
11028	KOENE	2	1998-04-16	1998-05-14	1998-04-22	1	29.59	Königlich Essen	Maubelstr. 90	Brandenburg	\N	14776	Germany
11029	CHOPS	4	1998-04-16	1998-05-14	1998-04-27	1	47.84	Chop-suey Chinese	Hauptstr. 31	Bern	\N	3012	Switzerland
11030	SAVEA	7	1998-04-17	1998-05-15	1998-04-27	2	830.75	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
11031	SAVEA	6	1998-04-17	1998-05-15	1998-04-24	2	227.22	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
11032	WHITC	2	1998-04-17	1998-05-15	1998-04-23	3	606.19	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
11033	RICSU	7	1998-04-17	1998-05-15	1998-04-23	3	84.74	Richter Supermarkt	Starenweg 5	Genève	\N	1204	Switzerland
11034	OLDWO	8	1998-04-20	1998-06-01	1998-04-27	1	40.32	Old World Delicatessen	2743 Bering St.	Anchorage	AK	99508	USA
11035	SUPRD	2	1998-04-20	1998-05-18	1998-04-24	2	0.17	Suprêmes délices	Boulevard Tirou, 255	Charleroi	\N	B-6000	Belgium
11036	DRACD	8	1998-04-20	1998-05-18	1998-04-22	3	149.47	Drachenblut Delikatessen	Walserweg 21	Aachen	\N	52066	Germany
11037	GODOS	7	1998-04-21	1998-05-19	1998-04-27	1	3.2	Godos Cocina Típica	C/ Romero, 33	Sevilla	\N	41101	Spain
11038	SUPRD	1	1998-04-21	1998-05-19	1998-04-30	2	29.59	Suprêmes délices	Boulevard Tirou, 255	Charleroi	\N	B-6000	Belgium
11039	LINOD	1	1998-04-21	1998-05-19	\N	2	65	LINO-Delicateses	Ave. 5 de Mayo Porlamar	I. de Margarita	Nueva Esparta	4980	Venezuela
11040	GREAL	4	1998-04-22	1998-05-20	\N	3	18.84	Great Lakes Food Market	2732 Baker Blvd.	Eugene	OR	97403	USA
11041	CHOPS	3	1998-04-22	1998-05-20	1998-04-28	2	48.22	Chop-suey Chinese	Hauptstr. 31	Bern	\N	3012	Switzerland
11042	COMMI	2	1998-04-22	1998-05-06	1998-05-01	1	29.99	Comércio Mineiro	Av. dos Lusíadas, 23	Sao Paulo	SP	05432-043	Brazil
11043	SPECD	5	1998-04-22	1998-05-20	1998-04-29	2	8.8	Spécialités du monde	25, rue Lauriston	Paris	\N	75016	France
11044	WOLZA	4	1998-04-23	1998-05-21	1998-05-01	1	8.72	Wolski Zajazd	ul. Filtrowa 68	Warszawa	\N	01-012	Poland
11045	BOTTM	6	1998-04-23	1998-05-21	\N	2	70.58	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
11046	WANDK	8	1998-04-23	1998-05-21	1998-04-24	2	71.64	Die Wandernde Kuh	Adenauerallee 900	Stuttgart	\N	70563	Germany
11047	EASTC	7	1998-04-24	1998-05-22	1998-05-01	3	46.62	Eastern Connection	35 King George	London	\N	WX3 6FW	UK
11048	BOTTM	7	1998-04-24	1998-05-22	1998-04-30	3	24.12	Bottom-Dollar Markets	23 Tsawassen Blvd.	Tsawassen	BC	T2F 8M4	Canada
11049	GOURL	3	1998-04-24	1998-05-22	1998-05-04	1	8.34	Gourmet Lanchonetes	Av. Brasil, 442	Campinas	SP	04876-786	Brazil
11050	FOLKO	8	1998-04-27	1998-05-25	1998-05-05	2	59.41	Folk och fä HB	Åkergatan 24	Bräcke	\N	S-844 67	Sweden
11051	LAMAI	7	1998-04-27	1998-05-25	\N	3	2.79	La maison d'Asie	1 rue Alsace-Lorraine	Toulouse	\N	31000	France
11052	HANAR	3	1998-04-27	1998-05-25	1998-05-01	1	67.26	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
11053	PICCO	2	1998-04-27	1998-05-25	1998-04-29	2	53.05	Piccolo und mehr	Geislweg 14	Salzburg	\N	5020	Austria
11054	CACTU	8	1998-04-28	1998-05-26	\N	1	0.33	Cactus Comidas para llevar	Cerrito 333	Buenos Aires	\N	1010	Argentina
11055	HILAA	7	1998-04-28	1998-05-26	1998-05-05	2	120.92	HILARION-Abastos	Carrera 22 con Ave. Carlos Soublette #8-35	San Cristóbal	Táchira	5022	Venezuela
11056	EASTC	8	1998-04-28	1998-05-12	1998-05-01	2	278.96	Eastern Connection	35 King George	London	\N	WX3 6FW	UK
11057	NORTS	3	1998-04-29	1998-05-27	1998-05-01	3	4.13	North/South	South House 300 Queensbridge	London	\N	SW7 1RZ	UK
11058	BLAUS	9	1998-04-29	1998-05-27	\N	3	31.14	Blauer See Delikatessen	Forsterstr. 57	Mannheim	\N	68306	Germany
11059	RICAR	2	1998-04-29	1998-06-10	\N	2	85.8	Ricardo Adocicados	Av. Copacabana, 267	Rio de Janeiro	RJ	02389-890	Brazil
11060	FRANS	2	1998-04-30	1998-05-28	1998-05-04	2	10.98	Franchi S.p.A.	Via Monte Bianco 34	Torino	\N	10100	Italy
11061	GREAL	4	1998-04-30	1998-06-11	\N	3	14.01	Great Lakes Food Market	2732 Baker Blvd.	Eugene	OR	97403	USA
11062	REGGC	4	1998-04-30	1998-05-28	\N	2	29.93	Reggiani Caseifici	Strada Provinciale 124	Reggio Emilia	\N	42100	Italy
11063	HUNGO	3	1998-04-30	1998-05-28	1998-05-06	2	81.73	Hungry Owl All-Night Grocers	8 Johnstown Road	Cork	Co. Cork	\N	Ireland
11064	SAVEA	1	1998-05-01	1998-05-29	1998-05-04	1	30.09	Save-a-lot Markets	187 Suffolk Ln.	Boise	ID	83720	USA
11065	LILAS	8	1998-05-01	1998-05-29	\N	1	12.91	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
11066	WHITC	7	1998-05-01	1998-05-29	1998-05-04	2	44.72	White Clover Markets	1029 - 12th Ave. S.	Seattle	WA	98124	USA
11067	DRACD	1	1998-05-04	1998-05-18	1998-05-06	2	7.98	Drachenblut Delikatessen	Walserweg 21	Aachen	\N	52066	Germany
11068	QUEEN	8	1998-05-04	1998-06-01	\N	2	81.75	Queen Cozinha	Alameda dos Canàrios, 891	Sao Paulo	SP	05487-020	Brazil
11069	TORTU	1	1998-05-04	1998-06-01	1998-05-06	2	15.67	Tortuga Restaurante	Avda. Azteca 123	México D.F.	\N	05033	Mexico
11070	LEHMS	2	1998-05-05	1998-06-02	\N	1	136	Lehmanns Marktstand	Magazinweg 7	Frankfurt a.M.	\N	60528	Germany
11071	LILAS	1	1998-05-05	1998-06-02	\N	1	0.93	LILA-Supermercado	Carrera 52 con Ave. Bolívar #65-98 Llano Largo	Barquisimeto	Lara	3508	Venezuela
11072	ERNSH	4	1998-05-05	1998-06-02	\N	2	258.64	Ernst Handel	Kirchgasse 6	Graz	\N	8010	Austria
11073	PERIC	2	1998-05-05	1998-06-02	\N	2	24.95	Pericles Comidas clásicas	Calle Dr. Jorge Cash 321	México D.F.	\N	05033	Mexico
11074	SIMOB	7	1998-05-06	1998-06-03	\N	2	18.44	Simons bistro	Vinbæltet 34	Kobenhavn	\N	1734	Denmark
11075	RICSU	8	1998-05-06	1998-06-03	\N	2	6.19	Richter Supermarkt	Starenweg 5	Genève	\N	1204	Switzerland
11076	BONAP	4	1998-05-06	1998-06-03	\N	2	38.28	Bon app'	12, rue des Bouchers	Marseille	\N	13008	France
11077	RATTC	1	1998-05-06	1998-06-03	\N	2	8.53	Rattlesnake Canyon Grocery	2817 Milton Dr.	Albuquerque	NM	87110	USA
10250	HANAR	4	1996-07-08	1996-08-05	1996-07-12	2	65.83	Hanari Carnes	Rua do Paço, 67	Rio de Janeiro	RJ	05454-876	Brazil
\.


--
-- Data for Name: products; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.products (product_id, product_name, supplier_id, category_id, quantity_per_unit, unit_price, units_in_stock, units_on_order, reorder_level, discontinued) FROM stdin;
1	Chai	8	1	10 boxes x 30 bags	18	39	0	10	1
2	Chang	1	1	24 - 12 oz bottles	19	17	40	25	1
3	Aniseed Syrup	1	2	12 - 550 ml bottles	10	13	70	25	0
4	Chef Anton's Cajun Seasoning	2	2	48 - 6 oz jars	22	53	0	0	0
5	Chef Anton's Gumbo Mix	2	2	36 boxes	21.35	0	0	0	1
6	Grandma's Boysenberry Spread	3	2	12 - 8 oz jars	25	120	0	25	0
7	Uncle Bob's Organic Dried Pears	3	7	12 - 1 lb pkgs.	30	15	0	10	0
8	Northwoods Cranberry Sauce	3	2	12 - 12 oz jars	40	6	0	0	0
9	Mishi Kobe Niku	4	6	18 - 500 g pkgs.	97	29	0	0	1
10	Ikura	4	8	12 - 200 ml jars	31	31	0	0	0
11	Queso Cabrales	5	4	1 kg pkg.	21	22	30	30	0
12	Queso Manchego La Pastora	5	4	10 - 500 g pkgs.	38	86	0	0	0
13	Konbu	6	8	2 kg box	6	24	0	5	0
14	Tofu	6	7	40 - 100 g pkgs.	23.25	35	0	0	0
15	Genen Shouyu	6	2	24 - 250 ml bottles	13	39	0	5	0
16	Pavlova	7	3	32 - 500 g boxes	17.45	29	0	10	0
17	Alice Mutton	7	6	20 - 1 kg tins	39	0	0	0	1
18	Carnarvon Tigers	7	8	16 kg pkg.	62.5	42	0	0	0
19	Teatime Chocolate Biscuits	8	3	10 boxes x 12 pieces	9.2	25	0	5	0
20	Sir Rodney's Marmalade	8	3	30 gift boxes	81	40	0	0	0
21	Sir Rodney's Scones	8	3	24 pkgs. x 4 pieces	10	3	40	5	0
22	Gustaf's Knäckebröd	9	5	24 - 500 g pkgs.	21	104	0	25	0
23	Tunnbröd	9	5	12 - 250 g pkgs.	9	61	0	25	0
24	Guaraná Fantástica	10	1	12 - 355 ml cans	4.5	20	0	0	1
25	NuNuCa Nuß-Nougat-Creme	11	3	20 - 450 g glasses	14	76	0	30	0
26	Gumbär Gummibärchen	11	3	100 - 250 g bags	31.23	15	0	0	0
27	Schoggi Schokolade	11	3	100 - 100 g pieces	43.9	49	0	30	0
28	Rössle Sauerkraut	12	7	25 - 825 g cans	45.6	26	0	0	1
29	Thüringer Rostbratwurst	12	6	50 bags x 30 sausgs.	123.79	0	0	0	1
30	Nord-Ost Matjeshering	13	8	10 - 200 g glasses	25.89	10	0	15	0
31	Gorgonzola Telino	14	4	12 - 100 g pkgs	12.5	0	70	20	0
32	Mascarpone Fabioli	14	4	24 - 200 g pkgs.	32	9	40	25	0
33	Geitost	15	4	500 g	2.5	112	0	20	0
34	Sasquatch Ale	16	1	24 - 12 oz bottles	14	111	0	15	0
35	Steeleye Stout	16	1	24 - 12 oz bottles	18	20	0	15	0
36	Inlagd Sill	17	8	24 - 250 g  jars	19	112	0	20	0
37	Gravad lax	17	8	12 - 500 g pkgs.	26	11	50	25	0
38	Côte de Blaye	18	1	12 - 75 cl bottles	263.5	17	0	15	0
39	Chartreuse verte	18	1	750 cc per bottle	18	69	0	5	0
40	Boston Crab Meat	19	8	24 - 4 oz tins	18.4	123	0	30	0
41	Jack's New England Clam Chowder	19	8	12 - 12 oz cans	9.65	85	0	10	0
42	Singaporean Hokkien Fried Mee	20	5	32 - 1 kg pkgs.	14	26	0	0	1
43	Ipoh Coffee	20	1	16 - 500 g tins	46	17	10	25	0
44	Gula Malacca	20	2	20 - 2 kg bags	19.45	27	0	15	0
45	Rogede sild	21	8	1k pkg.	9.5	5	70	15	0
46	Spegesild	21	8	4 - 450 g glasses	12	95	0	0	0
47	Zaanse koeken	22	3	10 - 4 oz boxes	9.5	36	0	0	0
48	Chocolade	22	3	10 pkgs.	12.75	15	70	25	0
49	Maxilaku	23	3	24 - 50 g pkgs.	20	10	60	15	0
50	Valkoinen suklaa	23	3	12 - 100 g bars	16.25	65	0	30	0
51	Manjimup Dried Apples	24	7	50 - 300 g pkgs.	53	20	0	10	0
52	Filo Mix	24	5	16 - 2 kg boxes	7	38	0	25	0
53	Perth Pasties	24	6	48 pieces	32.8	0	0	0	1
54	Tourtière	25	6	16 pies	7.45	21	0	10	0
55	Pâté chinois	25	6	24 boxes x 2 pies	24	115	0	20	0
56	Gnocchi di nonna Alice	26	5	24 - 250 g pkgs.	38	21	10	30	0
57	Ravioli Angelo	26	5	24 - 250 g pkgs.	19.5	36	0	20	0
58	Escargots de Bourgogne	27	8	24 pieces	13.25	62	0	20	0
59	Raclette Courdavault	28	4	5 kg pkg.	55	79	0	0	0
60	Camembert Pierrot	28	4	15 - 300 g rounds	34	19	0	0	0
61	Sirop d'érable	29	2	24 - 500 ml bottles	28.5	113	0	25	0
62	Tarte au sucre	29	3	48 pies	49.3	17	0	0	0
63	Vegie-spread	7	2	15 - 625 g jars	43.9	24	0	5	0
64	Wimmers gute Semmelknödel	12	5	20 bags x 4 pieces	33.25	22	80	30	0
65	Louisiana Fiery Hot Pepper Sauce	2	2	32 - 8 oz bottles	21.05	76	0	0	0
66	Louisiana Hot Spiced Okra	2	2	24 - 8 oz jars	17	4	100	20	0
67	Laughing Lumberjack Lager	16	1	24 - 12 oz bottles	14	52	0	10	0
68	Scottish Longbreads	8	3	10 boxes x 8 pieces	12.5	6	10	15	0
69	Gudbrandsdalsost	15	4	10 kg pkg.	36	26	0	15	0
70	Outback Lager	7	1	24 - 355 ml bottles	15	15	10	30	0
71	Flotemysost	15	4	10 - 500 g pkgs.	21.5	26	0	0	0
72	Mozzarella di Giovanni	14	4	24 - 200 g pkgs.	34.8	14	0	0	0
73	Röd Kaviar	17	8	24 - 150 g jars	15	101	0	5	0
74	Longlife Tofu	4	7	5 kg pkg.	10	4	20	5	0
75	Rhönbräu Klosterbier	12	1	24 - 0.5 l bottles	7.75	125	0	25	0
76	Lakkalikööri	23	1	500 ml	18	57	0	20	0
77	Original Frankfurter grüne Soße	12	2	12 boxes	13	32	0	15	0
\.


--
-- Data for Name: region; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.region (region_id, region_description) FROM stdin;
1	Eastern
2	Western
3	Northern
4	Southern
\.


--
-- Data for Name: shippers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.shippers (shipper_id, company_name, phone) FROM stdin;
1	Speedy Express	(503) 555-9831
2	United Package	(503) 555-3199
3	Federal Shipping	(503) 555-9931
4	Alliance Shippers	1-800-222-0451
5	UPS	1-800-782-7892
6	DHL	1-800-225-5345
\.


--
-- Data for Name: suppliers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.suppliers (supplier_id, company_name, contact_name, contact_title, address, city, region, postal_code, country, phone, fax, homepage) FROM stdin;
1	Exotic Liquids	Charlotte Cooper	Purchasing Manager	49 Gilbert St.	London	\N	EC1 4SD	UK	(171) 555-2222	\N	\N
2	New Orleans Cajun Delights	Shelley Burke	Order Administrator	P.O. Box 78934	New Orleans	LA	70117	USA	(100) 555-4822	\N	#CAJUN.HTM#
3	Grandma Kelly's Homestead	Regina Murphy	Sales Representative	707 Oxford Rd.	Ann Arbor	MI	48104	USA	(313) 555-5735	(313) 555-3349	\N
4	Tokyo Traders	Yoshi Nagase	Marketing Manager	9-8 Sekimai Musashino-shi	Tokyo	\N	100	Japan	(03) 3555-5011	\N	\N
5	Cooperativa de Quesos 'Las Cabras'	Antonio del Valle Saavedra	Export Administrator	Calle del Rosal 4	Oviedo	Asturias	33007	Spain	(98) 598 76 54	\N	\N
6	Mayumi's	Mayumi Ohno	Marketing Representative	92 Setsuko Chuo-ku	Osaka	\N	545	Japan	(06) 431-7877	\N	Mayumi's (on the World Wide Web)#http://www.microsoft.com/accessdev/sampleapps/mayumi.htm#
7	Pavlova, Ltd.	Ian Devling	Marketing Manager	74 Rose St. Moonie Ponds	Melbourne	Victoria	3058	Australia	(03) 444-2343	(03) 444-6588	\N
8	Specialty Biscuits, Ltd.	Peter Wilson	Sales Representative	29 King's Way	Manchester	\N	M14 GSD	UK	(161) 555-4448	\N	\N
9	PB Knäckebröd AB	Lars Peterson	Sales Agent	Kaloadagatan 13	Göteborg	\N	S-345 67	Sweden	031-987 65 43	031-987 65 91	\N
10	Refrescos Americanas LTDA	Carlos Diaz	Marketing Manager	Av. das Americanas 12.890	Sao Paulo	\N	5442	Brazil	(11) 555 4640	\N	\N
11	Heli Süßwaren GmbH & Co. KG	Petra Winkler	Sales Manager	Tiergartenstraße 5	Berlin	\N	10785	Germany	(010) 9984510	\N	\N
12	Plutzer Lebensmittelgroßmärkte AG	postgres Bein	International Marketing Mgr.	Bogenallee 51	Frankfurt	\N	60439	Germany	(069) 992755	\N	Plutzer (on the World Wide Web)#http://www.microsoft.com/accessdev/sampleapps/plutzer.htm#
13	Nord-Ost-Fisch Handelsgesellschaft mbH	Sven Petersen	Coordinator Foreign Markets	Frahmredder 112a	Cuxhaven	\N	27478	Germany	(04721) 8713	(04721) 8714	\N
14	Formaggi Fortini s.r.l.	Elio Rossi	Sales Representative	Viale Dante, 75	Ravenna	\N	48100	Italy	(0544) 60323	(0544) 60603	#FORMAGGI.HTM#
15	Norske Meierier	Beate Vileid	Marketing Manager	Hatlevegen 5	Sandvika	\N	1320	Norway	(0)2-953010	\N	\N
16	Bigfoot Breweries	Cheryl Saylor	Regional Account Rep.	3400 - 8th Avenue Suite 210	Bend	OR	97101	USA	(503) 555-9931	\N	\N
17	Svensk Sjöföda AB	Michael Björn	Sales Representative	Brovallavägen 231	Stockholm	\N	S-123 45	Sweden	08-123 45 67	\N	\N
18	Aux joyeux ecclésiastiques	Guylène Nodier	Sales Manager	203, Rue des Francs-Bourgeois	Paris	\N	75004	France	(1) 03.83.00.68	(1) 03.83.00.62	\N
19	New England Seafood Cannery	Robb Merchant	Wholesale Account Agent	Order Processing Dept. 2100 Paul Revere Blvd.	Boston	MA	02134	USA	(617) 555-3267	(617) 555-3389	\N
20	Leka Trading	Chandra Leka	Owner	471 Serangoon Loop, Suite #402	Singapore	\N	0512	Singapore	555-8787	\N	\N
21	Lyngbysild	Niels Petersen	Sales Manager	Lyngbysild Fiskebakken 10	Lyngby	\N	2800	Denmark	43844108	43844115	\N
22	Zaanse Snoepfabriek	Dirk Luchte	Accounting Manager	Verkoop Rijnweg 22	Zaandam	\N	9999 ZZ	Netherlands	(12345) 1212	(12345) 1210	\N
23	Karkki Oy	Anne Heikkonen	Product Manager	Valtakatu 12	Lappeenranta	\N	53120	Finland	(953) 10956	\N	\N
24	G'day, Mate	Wendy Mackenzie	Sales Representative	170 Prince Edward Parade Hunter's Hill	Sydney	NSW	2042	Australia	(02) 555-5914	(02) 555-4873	G'day Mate (on the World Wide Web)#http://www.microsoft.com/accessdev/sampleapps/gdaymate.htm#
25	Ma Maison	Jean-Guy Lauzon	Marketing Manager	2960 Rue St. Laurent	Montréal	Québec	H1J 1C3	Canada	(514) 555-9022	\N	\N
26	Pasta Buttini s.r.l.	Giovanni Giudici	Order Administrator	Via dei Gelsomini, 153	Salerno	\N	84100	Italy	(089) 6547665	(089) 6547667	\N
27	Escargots Nouveaux	Marie Delamare	Sales Manager	22, rue H. Voiron	Montceau	\N	71300	France	85.57.00.07	\N	\N
28	Gai pâturage	Eliane Noz	Sales Representative	Bat. B 3, rue des Alpes	Annecy	\N	74000	France	38.76.98.06	38.76.98.58	\N
29	Forêts d'érables	Chantal Goulet	Accounting Manager	148 rue Chasseur	Ste-Hyacinthe	Québec	J2S 7S8	Canada	(514) 555-2955	(514) 555-2921	\N
\.


--
-- Data for Name: territories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.territories (territory_id, territory_description, region_id) FROM stdin;
01581	Westboro	1
01730	Bedford	1
01833	Georgetow	1
02116	Boston	1
02139	Cambridge	1
02184	Braintree	1
02903	Providence	1
03049	Hollis	3
03801	Portsmouth	3
06897	Wilton	1
07960	Morristown	1
08837	Edison	1
10019	New York	1
10038	New York	1
11747	Mellvile	1
14450	Fairport	1
19428	Philadelphia	3
19713	Neward	1
20852	Rockville	1
27403	Greensboro	1
27511	Cary	1
29202	Columbia	4
30346	Atlanta	4
31406	Savannah	4
32859	Orlando	4
33607	Tampa	4
40222	Louisville	1
44122	Beachwood	3
45839	Findlay	3
48075	Southfield	3
48084	Troy	3
48304	Bloomfield Hills	3
53404	Racine	3
55113	Roseville	3
55439	Minneapolis	3
60179	Hoffman Estates	2
60601	Chicago	2
72716	Bentonville	4
75234	Dallas	4
78759	Austin	4
80202	Denver	2
80909	Colorado Springs	2
85014	Phoenix	2
85251	Scottsdale	2
90405	Santa Monica	2
94025	Menlo Park	2
94105	San Francisco	2
95008	Campbell	2
95054	Santa Clara	2
95060	Santa Cruz	2
98004	Bellevue	2
98052	Redmond	2
98104	Seattle	2
\.


--
-- Data for Name: us_states; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.us_states (state_id, state_name, state_abbr, state_region) FROM stdin;
1	Alabama	AL	south
2	Alaska	AK	north
3	Arizona	AZ	west
4	Arkansas	AR	south
5	California	CA	west
6	Colorado	CO	west
7	Connecticut	CT	east
8	Delaware	DE	east
9	District of Columbia	DC	east
10	Florida	FL	south
11	Georgia	GA	south
12	Hawaii	HI	west
13	Idaho	ID	midwest
14	Illinois	IL	midwest
15	Indiana	IN	midwest
16	Iowa	IO	midwest
17	Kansas	KS	midwest
18	Kentucky	KY	south
19	Louisiana	LA	south
20	Maine	ME	north
21	Maryland	MD	east
22	Massachusetts	MA	north
23	Michigan	MI	north
24	Minnesota	MN	north
25	Mississippi	MS	south
26	Missouri	MO	south
27	Montana	MT	west
28	Nebraska	NE	midwest
29	Nevada	NV	west
30	New Hampshire	NH	east
31	New Jersey	NJ	east
32	New Mexico	NM	west
33	New York	NY	east
34	North Carolina	NC	east
35	North Dakota	ND	midwest
36	Ohio	OH	midwest
37	Oklahoma	OK	midwest
38	Oregon	OR	west
39	Pennsylvania	PA	east
40	Rhode Island	RI	east
41	South Carolina	SC	east
42	South Dakota	SD	midwest
43	Tennessee	TN	midwest
44	Texas	TX	west
45	Utah	UT	west
46	Vermont	VT	east
47	Virginia	VA	east
48	Washington	WA	west
49	West Virginia	WV	south
50	Wisconsin	WI	midwest
51	Wyoming	WY	west
\.


--
-- Data for Name: xc_knex_migrations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.xc_knex_migrations (id, name, batch, migration_time) FROM stdin;
1	project	1	2022-05-18 11:22:25.307+02
2	m2m	1	2022-05-18 11:22:25.31+02
3	fkn	1	2022-05-18 11:22:25.311+02
4	viewType	1	2022-05-18 11:22:25.312+02
5	viewName	1	2022-05-18 11:22:25.313+02
6	nc_006_alter_nc_shared_views	1	2022-05-18 11:22:25.315+02
7	nc_007_alter_nc_shared_views_1	1	2022-05-18 11:22:25.316+02
8	nc_008_add_nc_shared_bases	1	2022-05-18 11:22:25.322+02
9	nc_009_add_model_order	1	2022-05-18 11:22:25.328+02
10	nc_010_add_parent_title_column	1	2022-05-18 11:22:25.329+02
11	nc_011_remove_old_ses_plugin	1	2022-05-18 11:22:25.33+02
\.


--
-- Data for Name: xc_knex_migrations_lock; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.xc_knex_migrations_lock (index, is_locked) FROM stdin;
1	0
\.


--
-- Data for Name: xc_knex_migrationsv2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.xc_knex_migrationsv2 (id, name, batch, migration_time) FROM stdin;
1	nc_011	1	2022-05-18 11:22:25.534+02
2	nc_012_alter_column_data_types	1	2022-05-18 11:22:25.542+02
\.


--
-- Data for Name: xc_knex_migrationsv2_lock; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.xc_knex_migrationsv2_lock (index, is_locked) FROM stdin;
1	0
\.


--
-- Name: nc_acl_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_acl_id_seq', 1, false);


--
-- Name: nc_api_tokens_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_api_tokens_id_seq', 1, false);


--
-- Name: nc_audit_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_audit_id_seq', 1, false);


--
-- Name: nc_cron_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_cron_id_seq', 1, false);


--
-- Name: nc_disabled_models_for_role_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_disabled_models_for_role_id_seq', 1, false);


--
-- Name: nc_evolutions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_evolutions_id_seq', 1, false);


--
-- Name: nc_hooks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_hooks_id_seq', 1, true);


--
-- Name: nc_loaders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_loaders_id_seq', 1, false);


--
-- Name: nc_migrations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_migrations_id_seq', 1, false);


--
-- Name: nc_models_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_models_id_seq', 1, false);


--
-- Name: nc_plugins_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_plugins_id_seq', 3, true);


--
-- Name: nc_relations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_relations_id_seq', 1, false);


--
-- Name: nc_resolvers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_resolvers_id_seq', 1, false);


--
-- Name: nc_roles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_roles_id_seq', 5, true);


--
-- Name: nc_routes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_routes_id_seq', 1, false);


--
-- Name: nc_rpc_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_rpc_id_seq', 1, false);


--
-- Name: nc_shared_bases_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_shared_bases_id_seq', 1, false);


--
-- Name: nc_shared_views_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_shared_views_id_seq', 1, false);


--
-- Name: nc_store_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.nc_store_id_seq', 5, true);


--
-- Name: xc_knex_migrations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.xc_knex_migrations_id_seq', 11, true);


--
-- Name: xc_knex_migrations_lock_index_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.xc_knex_migrations_lock_index_seq', 1, true);


--
-- Name: xc_knex_migrationsv2_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.xc_knex_migrationsv2_id_seq', 2, true);


--
-- Name: xc_knex_migrationsv2_lock_index_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.xc_knex_migrationsv2_lock_index_seq', 1, true);


--
-- Name: nc_acl nc_acl_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_acl
    ADD CONSTRAINT nc_acl_pkey PRIMARY KEY (id);


--
-- Name: nc_api_tokens nc_api_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_api_tokens
    ADD CONSTRAINT nc_api_tokens_pkey PRIMARY KEY (id);


--
-- Name: nc_audit nc_audit_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_audit
    ADD CONSTRAINT nc_audit_pkey PRIMARY KEY (id);


--
-- Name: nc_audit_v2 nc_audit_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_audit_v2
    ADD CONSTRAINT nc_audit_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_bases_v2 nc_bases_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_bases_v2
    ADD CONSTRAINT nc_bases_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_col_formula_v2 nc_col_formula_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_formula_v2
    ADD CONSTRAINT nc_col_formula_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_col_lookup_v2 nc_col_lookup_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_lookup_v2
    ADD CONSTRAINT nc_col_lookup_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_col_relations_v2 nc_col_relations_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_relations_v2
    ADD CONSTRAINT nc_col_relations_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_col_rollup_v2 nc_col_rollup_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_rollup_v2
    ADD CONSTRAINT nc_col_rollup_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_col_select_options_v2 nc_col_select_options_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_select_options_v2
    ADD CONSTRAINT nc_col_select_options_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_columns_v2 nc_columns_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_columns_v2
    ADD CONSTRAINT nc_columns_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_cron nc_cron_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_cron
    ADD CONSTRAINT nc_cron_pkey PRIMARY KEY (id);


--
-- Name: nc_disabled_models_for_role nc_disabled_models_for_role_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_disabled_models_for_role
    ADD CONSTRAINT nc_disabled_models_for_role_pkey PRIMARY KEY (id);


--
-- Name: nc_disabled_models_for_role_v2 nc_disabled_models_for_role_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_disabled_models_for_role_v2
    ADD CONSTRAINT nc_disabled_models_for_role_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_evolutions nc_evolutions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_evolutions
    ADD CONSTRAINT nc_evolutions_pkey PRIMARY KEY (id);


--
-- Name: nc_filter_exp_v2 nc_filter_exp_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_filter_exp_v2
    ADD CONSTRAINT nc_filter_exp_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_form_view_columns_v2 nc_form_view_columns_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_form_view_columns_v2
    ADD CONSTRAINT nc_form_view_columns_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_form_view_v2 nc_form_view_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_form_view_v2
    ADD CONSTRAINT nc_form_view_v2_pkey PRIMARY KEY (fk_view_id);


--
-- Name: nc_gallery_view_columns_v2 nc_gallery_view_columns_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_gallery_view_columns_v2
    ADD CONSTRAINT nc_gallery_view_columns_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_gallery_view_v2 nc_gallery_view_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_gallery_view_v2
    ADD CONSTRAINT nc_gallery_view_v2_pkey PRIMARY KEY (fk_view_id);


--
-- Name: nc_grid_view_columns_v2 nc_grid_view_columns_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_grid_view_columns_v2
    ADD CONSTRAINT nc_grid_view_columns_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_grid_view_v2 nc_grid_view_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_grid_view_v2
    ADD CONSTRAINT nc_grid_view_v2_pkey PRIMARY KEY (fk_view_id);


--
-- Name: nc_hook_logs_v2 nc_hook_logs_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_hook_logs_v2
    ADD CONSTRAINT nc_hook_logs_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_hooks nc_hooks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_hooks
    ADD CONSTRAINT nc_hooks_pkey PRIMARY KEY (id);


--
-- Name: nc_hooks_v2 nc_hooks_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_hooks_v2
    ADD CONSTRAINT nc_hooks_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_kanban_view_columns_v2 nc_kanban_view_columns_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_kanban_view_columns_v2
    ADD CONSTRAINT nc_kanban_view_columns_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_kanban_view_v2 nc_kanban_view_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_kanban_view_v2
    ADD CONSTRAINT nc_kanban_view_v2_pkey PRIMARY KEY (fk_view_id);


--
-- Name: nc_loaders nc_loaders_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_loaders
    ADD CONSTRAINT nc_loaders_pkey PRIMARY KEY (id);


--
-- Name: nc_migrations nc_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_migrations
    ADD CONSTRAINT nc_migrations_pkey PRIMARY KEY (id);


--
-- Name: nc_models nc_models_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_models
    ADD CONSTRAINT nc_models_pkey PRIMARY KEY (id);


--
-- Name: nc_models_v2 nc_models_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_models_v2
    ADD CONSTRAINT nc_models_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_orgs_v2 nc_orgs_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_orgs_v2
    ADD CONSTRAINT nc_orgs_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_plugins nc_plugins_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_plugins
    ADD CONSTRAINT nc_plugins_pkey PRIMARY KEY (id);


--
-- Name: nc_plugins_v2 nc_plugins_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_plugins_v2
    ADD CONSTRAINT nc_plugins_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_projects nc_projects_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_projects
    ADD CONSTRAINT nc_projects_pkey PRIMARY KEY (id);


--
-- Name: nc_projects_v2 nc_projects_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_projects_v2
    ADD CONSTRAINT nc_projects_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_relations nc_relations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_relations
    ADD CONSTRAINT nc_relations_pkey PRIMARY KEY (id);


--
-- Name: nc_resolvers nc_resolvers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_resolvers
    ADD CONSTRAINT nc_resolvers_pkey PRIMARY KEY (id);


--
-- Name: nc_roles nc_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_roles
    ADD CONSTRAINT nc_roles_pkey PRIMARY KEY (id);


--
-- Name: nc_routes nc_routes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_routes
    ADD CONSTRAINT nc_routes_pkey PRIMARY KEY (id);


--
-- Name: nc_rpc nc_rpc_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_rpc
    ADD CONSTRAINT nc_rpc_pkey PRIMARY KEY (id);


--
-- Name: nc_shared_bases nc_shared_bases_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_shared_bases
    ADD CONSTRAINT nc_shared_bases_pkey PRIMARY KEY (id);


--
-- Name: nc_shared_views nc_shared_views_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_shared_views
    ADD CONSTRAINT nc_shared_views_pkey PRIMARY KEY (id);


--
-- Name: nc_shared_views_v2 nc_shared_views_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_shared_views_v2
    ADD CONSTRAINT nc_shared_views_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_sort_v2 nc_sort_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_sort_v2
    ADD CONSTRAINT nc_sort_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_store nc_store_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_store
    ADD CONSTRAINT nc_store_pkey PRIMARY KEY (id);


--
-- Name: nc_teams_v2 nc_teams_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_teams_v2
    ADD CONSTRAINT nc_teams_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_users_v2 nc_users_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_users_v2
    ADD CONSTRAINT nc_users_v2_pkey PRIMARY KEY (id);


--
-- Name: nc_views_v2 nc_views_v2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_views_v2
    ADD CONSTRAINT nc_views_v2_pkey PRIMARY KEY (id);


--
-- Name: categories pk_categories; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT pk_categories PRIMARY KEY (category_id);


--
-- Name: customer_customer_demo pk_customer_customer_demo; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.customer_customer_demo
    ADD CONSTRAINT pk_customer_customer_demo PRIMARY KEY (customer_id, customer_type_id);


--
-- Name: customer_demographics pk_customer_demographics; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.customer_demographics
    ADD CONSTRAINT pk_customer_demographics PRIMARY KEY (customer_type_id);


--
-- Name: customers pk_customers; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT pk_customers PRIMARY KEY (customer_id);


--
-- Name: employee_territories pk_employee_territories; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employee_territories
    ADD CONSTRAINT pk_employee_territories PRIMARY KEY (employee_id, territory_id);


--
-- Name: employees pk_employees; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT pk_employees PRIMARY KEY (employee_id);


--
-- Name: order_details pk_order_details; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_details
    ADD CONSTRAINT pk_order_details PRIMARY KEY (order_id, product_id);


--
-- Name: orders pk_orders; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT pk_orders PRIMARY KEY (order_id);


--
-- Name: products pk_products; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT pk_products PRIMARY KEY (product_id);


--
-- Name: region pk_region; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.region
    ADD CONSTRAINT pk_region PRIMARY KEY (region_id);


--
-- Name: shippers pk_shippers; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shippers
    ADD CONSTRAINT pk_shippers PRIMARY KEY (shipper_id);


--
-- Name: suppliers pk_suppliers; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.suppliers
    ADD CONSTRAINT pk_suppliers PRIMARY KEY (supplier_id);


--
-- Name: territories pk_territories; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.territories
    ADD CONSTRAINT pk_territories PRIMARY KEY (territory_id);


--
-- Name: us_states pk_usstates; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.us_states
    ADD CONSTRAINT pk_usstates PRIMARY KEY (state_id);


--
-- Name: xc_knex_migrations_lock xc_knex_migrations_lock_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xc_knex_migrations_lock
    ADD CONSTRAINT xc_knex_migrations_lock_pkey PRIMARY KEY (index);


--
-- Name: xc_knex_migrations xc_knex_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xc_knex_migrations
    ADD CONSTRAINT xc_knex_migrations_pkey PRIMARY KEY (id);


--
-- Name: xc_knex_migrationsv2_lock xc_knex_migrationsv2_lock_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xc_knex_migrationsv2_lock
    ADD CONSTRAINT xc_knex_migrationsv2_lock_pkey PRIMARY KEY (index);


--
-- Name: xc_knex_migrationsv2 xc_knex_migrationsv2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.xc_knex_migrationsv2
    ADD CONSTRAINT xc_knex_migrationsv2_pkey PRIMARY KEY (id);


--
-- Name: `nc_audit_index`; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "`nc_audit_index`" ON public.nc_audit USING btree (db_alias, project_id, model_name, model_id);


--
-- Name: nc_audit_v2_row_id_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX nc_audit_v2_row_id_index ON public.nc_audit_v2 USING btree (row_id);


--
-- Name: nc_models_db_alias_title_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX nc_models_db_alias_title_index ON public.nc_models USING btree (db_alias, title);


--
-- Name: nc_models_order_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX nc_models_order_index ON public.nc_models USING btree ("order");


--
-- Name: nc_models_view_order_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX nc_models_view_order_index ON public.nc_models USING btree (view_order);


--
-- Name: nc_projects_users_project_id_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX nc_projects_users_project_id_index ON public.nc_projects_users USING btree (project_id);


--
-- Name: nc_projects_users_user_id_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX nc_projects_users_user_id_index ON public.nc_projects_users USING btree (user_id);


--
-- Name: nc_relations_db_alias_tn_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX nc_relations_db_alias_tn_index ON public.nc_relations USING btree (db_alias, tn);


--
-- Name: nc_routes_db_alias_title_tn_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX nc_routes_db_alias_title_tn_index ON public.nc_routes USING btree (db_alias, title, tn);


--
-- Name: nc_store_key_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX nc_store_key_index ON public.nc_store USING btree (key);


--
-- Name: xc_disabled124_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX xc_disabled124_idx ON public.nc_disabled_models_for_role USING btree (project_id, db_alias, title, type, role);


--
-- Name: customer_customer_demo fk_customer_customer_demo_customer_demographics; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.customer_customer_demo
    ADD CONSTRAINT fk_customer_customer_demo_customer_demographics FOREIGN KEY (customer_type_id) REFERENCES public.customer_demographics(customer_type_id);


--
-- Name: customer_customer_demo fk_customer_customer_demo_customers; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.customer_customer_demo
    ADD CONSTRAINT fk_customer_customer_demo_customers FOREIGN KEY (customer_id) REFERENCES public.customers(customer_id);


--
-- Name: employee_territories fk_employee_territories_employees; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employee_territories
    ADD CONSTRAINT fk_employee_territories_employees FOREIGN KEY (employee_id) REFERENCES public.employees(employee_id);


--
-- Name: employee_territories fk_employee_territories_territories; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employee_territories
    ADD CONSTRAINT fk_employee_territories_territories FOREIGN KEY (territory_id) REFERENCES public.territories(territory_id);


--
-- Name: employees fk_employees_employees; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT fk_employees_employees FOREIGN KEY (reports_to) REFERENCES public.employees(employee_id);


--
-- Name: order_details fk_order_details_orders; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_details
    ADD CONSTRAINT fk_order_details_orders FOREIGN KEY (order_id) REFERENCES public.orders(order_id);


--
-- Name: order_details fk_order_details_products; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_details
    ADD CONSTRAINT fk_order_details_products FOREIGN KEY (product_id) REFERENCES public.products(product_id);


--
-- Name: orders fk_orders_customers; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT fk_orders_customers FOREIGN KEY (customer_id) REFERENCES public.customers(customer_id);


--
-- Name: orders fk_orders_employees; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT fk_orders_employees FOREIGN KEY (employee_id) REFERENCES public.employees(employee_id);


--
-- Name: orders fk_orders_shippers; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT fk_orders_shippers FOREIGN KEY (ship_via) REFERENCES public.shippers(shipper_id);


--
-- Name: products fk_products_categories; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT fk_products_categories FOREIGN KEY (category_id) REFERENCES public.categories(category_id);


--
-- Name: products fk_products_suppliers; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT fk_products_suppliers FOREIGN KEY (supplier_id) REFERENCES public.suppliers(supplier_id);


--
-- Name: territories fk_territories_region; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.territories
    ADD CONSTRAINT fk_territories_region FOREIGN KEY (region_id) REFERENCES public.region(region_id);


--
-- Name: nc_audit_v2 nc_audit_v2_base_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_audit_v2
    ADD CONSTRAINT nc_audit_v2_base_id_foreign FOREIGN KEY (base_id) REFERENCES public.nc_bases_v2(id);


--
-- Name: nc_audit_v2 nc_audit_v2_fk_model_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_audit_v2
    ADD CONSTRAINT nc_audit_v2_fk_model_id_foreign FOREIGN KEY (fk_model_id) REFERENCES public.nc_models_v2(id);


--
-- Name: nc_audit_v2 nc_audit_v2_project_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_audit_v2
    ADD CONSTRAINT nc_audit_v2_project_id_foreign FOREIGN KEY (project_id) REFERENCES public.nc_projects_v2(id);


--
-- Name: nc_bases_v2 nc_bases_v2_project_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_bases_v2
    ADD CONSTRAINT nc_bases_v2_project_id_foreign FOREIGN KEY (project_id) REFERENCES public.nc_projects_v2(id);


--
-- Name: nc_col_formula_v2 nc_col_formula_v2_fk_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_formula_v2
    ADD CONSTRAINT nc_col_formula_v2_fk_column_id_foreign FOREIGN KEY (fk_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_col_lookup_v2 nc_col_lookup_v2_fk_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_lookup_v2
    ADD CONSTRAINT nc_col_lookup_v2_fk_column_id_foreign FOREIGN KEY (fk_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_col_lookup_v2 nc_col_lookup_v2_fk_lookup_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_lookup_v2
    ADD CONSTRAINT nc_col_lookup_v2_fk_lookup_column_id_foreign FOREIGN KEY (fk_lookup_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_col_lookup_v2 nc_col_lookup_v2_fk_relation_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_lookup_v2
    ADD CONSTRAINT nc_col_lookup_v2_fk_relation_column_id_foreign FOREIGN KEY (fk_relation_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_col_relations_v2 nc_col_relations_v2_fk_child_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_relations_v2
    ADD CONSTRAINT nc_col_relations_v2_fk_child_column_id_foreign FOREIGN KEY (fk_child_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_col_relations_v2 nc_col_relations_v2_fk_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_relations_v2
    ADD CONSTRAINT nc_col_relations_v2_fk_column_id_foreign FOREIGN KEY (fk_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_col_relations_v2 nc_col_relations_v2_fk_mm_child_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_relations_v2
    ADD CONSTRAINT nc_col_relations_v2_fk_mm_child_column_id_foreign FOREIGN KEY (fk_mm_child_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_col_relations_v2 nc_col_relations_v2_fk_mm_model_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_relations_v2
    ADD CONSTRAINT nc_col_relations_v2_fk_mm_model_id_foreign FOREIGN KEY (fk_mm_model_id) REFERENCES public.nc_models_v2(id);


--
-- Name: nc_col_relations_v2 nc_col_relations_v2_fk_mm_parent_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_relations_v2
    ADD CONSTRAINT nc_col_relations_v2_fk_mm_parent_column_id_foreign FOREIGN KEY (fk_mm_parent_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_col_relations_v2 nc_col_relations_v2_fk_parent_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_relations_v2
    ADD CONSTRAINT nc_col_relations_v2_fk_parent_column_id_foreign FOREIGN KEY (fk_parent_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_col_relations_v2 nc_col_relations_v2_fk_related_model_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_relations_v2
    ADD CONSTRAINT nc_col_relations_v2_fk_related_model_id_foreign FOREIGN KEY (fk_related_model_id) REFERENCES public.nc_models_v2(id);


--
-- Name: nc_col_rollup_v2 nc_col_rollup_v2_fk_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_rollup_v2
    ADD CONSTRAINT nc_col_rollup_v2_fk_column_id_foreign FOREIGN KEY (fk_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_col_rollup_v2 nc_col_rollup_v2_fk_relation_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_rollup_v2
    ADD CONSTRAINT nc_col_rollup_v2_fk_relation_column_id_foreign FOREIGN KEY (fk_relation_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_col_rollup_v2 nc_col_rollup_v2_fk_rollup_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_rollup_v2
    ADD CONSTRAINT nc_col_rollup_v2_fk_rollup_column_id_foreign FOREIGN KEY (fk_rollup_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_col_select_options_v2 nc_col_select_options_v2_fk_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_col_select_options_v2
    ADD CONSTRAINT nc_col_select_options_v2_fk_column_id_foreign FOREIGN KEY (fk_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_columns_v2 nc_columns_v2_fk_model_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_columns_v2
    ADD CONSTRAINT nc_columns_v2_fk_model_id_foreign FOREIGN KEY (fk_model_id) REFERENCES public.nc_models_v2(id);


--
-- Name: nc_disabled_models_for_role_v2 nc_disabled_models_for_role_v2_fk_view_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_disabled_models_for_role_v2
    ADD CONSTRAINT nc_disabled_models_for_role_v2_fk_view_id_foreign FOREIGN KEY (fk_view_id) REFERENCES public.nc_views_v2(id);


--
-- Name: nc_filter_exp_v2 nc_filter_exp_v2_fk_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_filter_exp_v2
    ADD CONSTRAINT nc_filter_exp_v2_fk_column_id_foreign FOREIGN KEY (fk_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_filter_exp_v2 nc_filter_exp_v2_fk_hook_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_filter_exp_v2
    ADD CONSTRAINT nc_filter_exp_v2_fk_hook_id_foreign FOREIGN KEY (fk_hook_id) REFERENCES public.nc_hooks_v2(id);


--
-- Name: nc_filter_exp_v2 nc_filter_exp_v2_fk_parent_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_filter_exp_v2
    ADD CONSTRAINT nc_filter_exp_v2_fk_parent_id_foreign FOREIGN KEY (fk_parent_id) REFERENCES public.nc_filter_exp_v2(id);


--
-- Name: nc_filter_exp_v2 nc_filter_exp_v2_fk_view_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_filter_exp_v2
    ADD CONSTRAINT nc_filter_exp_v2_fk_view_id_foreign FOREIGN KEY (fk_view_id) REFERENCES public.nc_views_v2(id);


--
-- Name: nc_form_view_columns_v2 nc_form_view_columns_v2_fk_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_form_view_columns_v2
    ADD CONSTRAINT nc_form_view_columns_v2_fk_column_id_foreign FOREIGN KEY (fk_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_form_view_columns_v2 nc_form_view_columns_v2_fk_view_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_form_view_columns_v2
    ADD CONSTRAINT nc_form_view_columns_v2_fk_view_id_foreign FOREIGN KEY (fk_view_id) REFERENCES public.nc_form_view_v2(fk_view_id);


--
-- Name: nc_form_view_v2 nc_form_view_v2_fk_view_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_form_view_v2
    ADD CONSTRAINT nc_form_view_v2_fk_view_id_foreign FOREIGN KEY (fk_view_id) REFERENCES public.nc_views_v2(id);


--
-- Name: nc_gallery_view_columns_v2 nc_gallery_view_columns_v2_fk_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_gallery_view_columns_v2
    ADD CONSTRAINT nc_gallery_view_columns_v2_fk_column_id_foreign FOREIGN KEY (fk_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_gallery_view_columns_v2 nc_gallery_view_columns_v2_fk_view_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_gallery_view_columns_v2
    ADD CONSTRAINT nc_gallery_view_columns_v2_fk_view_id_foreign FOREIGN KEY (fk_view_id) REFERENCES public.nc_gallery_view_v2(fk_view_id);


--
-- Name: nc_gallery_view_v2 nc_gallery_view_v2_fk_cover_image_col_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_gallery_view_v2
    ADD CONSTRAINT nc_gallery_view_v2_fk_cover_image_col_id_foreign FOREIGN KEY (fk_cover_image_col_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_gallery_view_v2 nc_gallery_view_v2_fk_view_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_gallery_view_v2
    ADD CONSTRAINT nc_gallery_view_v2_fk_view_id_foreign FOREIGN KEY (fk_view_id) REFERENCES public.nc_views_v2(id);


--
-- Name: nc_grid_view_columns_v2 nc_grid_view_columns_v2_fk_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_grid_view_columns_v2
    ADD CONSTRAINT nc_grid_view_columns_v2_fk_column_id_foreign FOREIGN KEY (fk_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_grid_view_columns_v2 nc_grid_view_columns_v2_fk_view_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_grid_view_columns_v2
    ADD CONSTRAINT nc_grid_view_columns_v2_fk_view_id_foreign FOREIGN KEY (fk_view_id) REFERENCES public.nc_grid_view_v2(fk_view_id);


--
-- Name: nc_grid_view_v2 nc_grid_view_v2_fk_view_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_grid_view_v2
    ADD CONSTRAINT nc_grid_view_v2_fk_view_id_foreign FOREIGN KEY (fk_view_id) REFERENCES public.nc_views_v2(id);


--
-- Name: nc_hooks_v2 nc_hooks_v2_fk_model_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_hooks_v2
    ADD CONSTRAINT nc_hooks_v2_fk_model_id_foreign FOREIGN KEY (fk_model_id) REFERENCES public.nc_models_v2(id);


--
-- Name: nc_kanban_view_columns_v2 nc_kanban_view_columns_v2_fk_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_kanban_view_columns_v2
    ADD CONSTRAINT nc_kanban_view_columns_v2_fk_column_id_foreign FOREIGN KEY (fk_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_kanban_view_columns_v2 nc_kanban_view_columns_v2_fk_view_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_kanban_view_columns_v2
    ADD CONSTRAINT nc_kanban_view_columns_v2_fk_view_id_foreign FOREIGN KEY (fk_view_id) REFERENCES public.nc_kanban_view_v2(fk_view_id);


--
-- Name: nc_kanban_view_v2 nc_kanban_view_v2_fk_view_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_kanban_view_v2
    ADD CONSTRAINT nc_kanban_view_v2_fk_view_id_foreign FOREIGN KEY (fk_view_id) REFERENCES public.nc_views_v2(id);


--
-- Name: nc_models_v2 nc_models_v2_base_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_models_v2
    ADD CONSTRAINT nc_models_v2_base_id_foreign FOREIGN KEY (base_id) REFERENCES public.nc_bases_v2(id);


--
-- Name: nc_models_v2 nc_models_v2_project_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_models_v2
    ADD CONSTRAINT nc_models_v2_project_id_foreign FOREIGN KEY (project_id) REFERENCES public.nc_projects_v2(id);


--
-- Name: nc_project_users_v2 nc_project_users_v2_fk_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_project_users_v2
    ADD CONSTRAINT nc_project_users_v2_fk_user_id_foreign FOREIGN KEY (fk_user_id) REFERENCES public.nc_users_v2(id);


--
-- Name: nc_project_users_v2 nc_project_users_v2_project_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_project_users_v2
    ADD CONSTRAINT nc_project_users_v2_project_id_foreign FOREIGN KEY (project_id) REFERENCES public.nc_projects_v2(id);


--
-- Name: nc_shared_views_v2 nc_shared_views_v2_fk_view_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_shared_views_v2
    ADD CONSTRAINT nc_shared_views_v2_fk_view_id_foreign FOREIGN KEY (fk_view_id) REFERENCES public.nc_views_v2(id);


--
-- Name: nc_sort_v2 nc_sort_v2_fk_column_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_sort_v2
    ADD CONSTRAINT nc_sort_v2_fk_column_id_foreign FOREIGN KEY (fk_column_id) REFERENCES public.nc_columns_v2(id);


--
-- Name: nc_sort_v2 nc_sort_v2_fk_view_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_sort_v2
    ADD CONSTRAINT nc_sort_v2_fk_view_id_foreign FOREIGN KEY (fk_view_id) REFERENCES public.nc_views_v2(id);


--
-- Name: nc_team_users_v2 nc_team_users_v2_org_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_team_users_v2
    ADD CONSTRAINT nc_team_users_v2_org_id_foreign FOREIGN KEY (org_id) REFERENCES public.nc_orgs_v2(id);


--
-- Name: nc_team_users_v2 nc_team_users_v2_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_team_users_v2
    ADD CONSTRAINT nc_team_users_v2_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.nc_users_v2(id);


--
-- Name: nc_teams_v2 nc_teams_v2_org_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_teams_v2
    ADD CONSTRAINT nc_teams_v2_org_id_foreign FOREIGN KEY (org_id) REFERENCES public.nc_orgs_v2(id);


--
-- Name: nc_views_v2 nc_views_v2_fk_model_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nc_views_v2
    ADD CONSTRAINT nc_views_v2_fk_model_id_foreign FOREIGN KEY (fk_model_id) REFERENCES public.nc_models_v2(id);


--
-- PostgreSQL database dump complete
--

