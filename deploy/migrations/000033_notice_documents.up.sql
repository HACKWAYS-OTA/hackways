ALTER TABLE site_notices
 ADD COLUMN document JSONB,
 ADD COLUMN body_format TEXT NOT NULL DEFAULT 'markdown' CHECK (body_format IN ('markdown','richtext')),
 ADD CONSTRAINT site_notices_document_check CHECK (
   (body_format='markdown' AND document IS NULL)
   OR (body_format='richtext' AND document IS NOT NULL AND jsonb_typeof(document)='object' AND octet_length(document::text)<=1048576)
 );
ALTER TABLE site_notices DROP CONSTRAINT site_notices_title_check;
ALTER TABLE site_notices ADD CONSTRAINT site_notices_title_check
 CHECK (char_length(title)<=180 AND (status<>'published' OR char_length(btrim(title))>0));
