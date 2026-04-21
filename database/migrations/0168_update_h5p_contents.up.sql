ALTER TABLE h5p_contents
    ADD COLUMN content_id varchar(255),
    ADD COLUMN library text;

ALTER TABLE h5p_contents
ADD CONSTRAINT h5p_contents_content_id_unique UNIQUE (content_id);

ALTER TABLE h5p_contents
ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
