UPDATE medias
SET static_url = regexp_replace(static_url, '^https://Clever Schoolx\.urp\.vn/', '')
WHERE static_url LIKE 'https://Clever Schoolx.urp.vn/%';
