CREATE TABLE programs (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 slug TEXT NOT NULL UNIQUE CHECK (slug ~ '^[a-z][a-z0-9-]{1,62}[a-z0-9]$'),
 title TEXT NOT NULL CHECK (length(title) BETWEEN 3 AND 120),
 summary TEXT NOT NULL CHECK (length(summary) BETWEEN 10 AND 400),
 overview TEXT NOT NULL CHECK (length(overview) BETWEEN 10 AND 6000),
 responsibilities TEXT NOT NULL CHECK (length(responsibilities) BETWEEN 10 AND 6000),
 recognition TEXT NOT NULL CHECK (length(recognition) BETWEEN 10 AND 6000),
 status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','open','closed','archived')),
 minimum_age INTEGER NOT NULL DEFAULT 18 CHECK (minimum_age BETWEEN 18 AND 100),
 questions JSONB NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(questions)='array' AND jsonb_array_length(questions)<=10),
 opens_at TIMESTAMPTZ, closes_at TIMESTAMPTZ,
 revision BIGINT NOT NULL DEFAULT 1, review_revision BIGINT NOT NULL DEFAULT 1,
 created_by UUID REFERENCES users(id),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(), updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 CHECK (opens_at IS NULL OR closes_at IS NULL OR closes_at>opens_at)
);
CREATE TABLE program_applications (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 program_id UUID NOT NULL REFERENCES programs(id),user_id UUID NOT NULL REFERENCES users(id),
 full_name TEXT NOT NULL DEFAULT '' CHECK (length(full_name)<=120),
 city TEXT NOT NULL DEFAULT '' CHECK (length(city)<=150),
 phone TEXT NOT NULL DEFAULT '' CHECK (length(phone)<=32),
 biography TEXT NOT NULL DEFAULT '' CHECK (length(biography)<=2000),
 profile_url TEXT NOT NULL DEFAULT '' CHECK (length(profile_url)<=2048),
 proposal TEXT NOT NULL DEFAULT '' CHECK (length(proposal)<=5000),
 age_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
 age INTEGER NOT NULL DEFAULT 0 CHECK(age BETWEEN 0 AND 120),
 answers JSONB NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(answers)='object'),
 question_snapshot JSONB NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(question_snapshot)='array'),
 status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','submitted','withdrawn')),
 decision TEXT NOT NULL DEFAULT 'pending' CHECK (decision IN ('pending','accepted','rejected','waitlisted')),
 reviewer_notes TEXT NOT NULL DEFAULT '' CHECK (length(reviewer_notes)<=3000),
 reviewed_by UUID REFERENCES users(id),reviewed_at TIMESTAMPTZ,
 published_decision TEXT CHECK (published_decision IN ('accepted','rejected','waitlisted')),
 published_at TIMESTAMPTZ,
 revision BIGINT NOT NULL DEFAULT 1,
 submitted_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 UNIQUE(program_id,user_id),
 CHECK ((published_at IS NULL)=(published_decision IS NULL)),
 CHECK (status<>'submitted' OR (age_confirmed AND length(full_name)>=2 AND length(city)>=2 AND length(proposal)>=30))
);
CREATE INDEX program_applications_review_idx ON program_applications(program_id,status,decision,created_at,id);
CREATE INDEX program_applications_user_idx ON program_applications(user_id,updated_at DESC,id);
CREATE TABLE program_releases (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),program_id UUID NOT NULL REFERENCES programs(id),
 request_key TEXT NOT NULL CHECK(length(request_key) BETWEEN 16 AND 128),
 payload_hash TEXT NOT NULL CHECK(length(payload_hash)=64),
 application_count INTEGER NOT NULL CHECK(application_count>0),
 released_by UUID NOT NULL REFERENCES users(id),created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 UNIQUE(program_id,request_key)
);
CREATE TABLE program_activity (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 program_id UUID NOT NULL REFERENCES programs(id),application_id UUID REFERENCES program_applications(id),
 actor_id UUID NOT NULL REFERENCES users(id),action TEXT NOT NULL CHECK(length(action)<=80),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX program_activity_program_idx ON program_activity(program_id,created_at DESC,id);
INSERT INTO programs(slug,title,summary,overview,responsibilities,recognition,status)
VALUES ('ambassadors','Hackways Ambassadors','Bring your community together. Open to anyone aged 18 or older.',
 'Become a Hackways ambassador and turn a shared interest into a gathering worth showing up for.',
 'Host your events using Hackways. Create an event page, share the invitation and manage your registrations through the platform.',
 'A letter of appreciation recognizes your work. The highest contributors receive swag. No additional benefits are currently announced.','open');
