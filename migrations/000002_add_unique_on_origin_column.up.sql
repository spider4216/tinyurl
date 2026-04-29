ALTER TABLE urls 
ADD CONSTRAINT origin_unique_idx UNIQUE (original);