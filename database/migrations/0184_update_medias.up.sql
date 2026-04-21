UPDATE medias
SET disk_name = 's3'
WHERE disk_name = 'public'
  AND type = 'file';
