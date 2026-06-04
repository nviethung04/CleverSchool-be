UPDATE medias
SET static_url = regexp_replace(static_url, '^https://lmsx\.urp\.vn/', '')
WHERE static_url LIKE 'https://lmsx.urp.vn/%';
