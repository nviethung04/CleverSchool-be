UPDATE medias
SET disk_name = 'public'
WHERE disk_name = 's3'
  AND type = 'file';
