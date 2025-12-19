-- Add cache configuration columns to proxy_hosts table
ALTER TABLE public.proxy_hosts ADD COLUMN IF NOT EXISTS cache_static_only boolean DEFAULT true NOT NULL;
ALTER TABLE public.proxy_hosts ADD COLUMN IF NOT EXISTS cache_ttl character varying(20) DEFAULT '7d'::character varying NOT NULL;

COMMENT ON COLUMN public.proxy_hosts.cache_static_only IS 'Only cache static assets (js, css, images, fonts) - excludes API paths';
COMMENT ON COLUMN public.proxy_hosts.cache_ttl IS 'Cache duration for static assets (e.g., "1h", "7d", "30m")';
