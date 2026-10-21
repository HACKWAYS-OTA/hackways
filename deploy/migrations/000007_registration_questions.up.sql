ALTER TABLE events
 ADD COLUMN registration_form_revision BIGINT NOT NULL DEFAULT 1 CHECK(registration_form_revision > 0),
 ADD COLUMN registration_name_required BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE event_registration_questions (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 public_code TEXT NOT NULL UNIQUE DEFAULT ('Q-' || encode(public.gen_random_bytes(8),'hex')),
 event_id UUID NOT NULL REFERENCES events(id),
 label TEXT NOT NULL CHECK(char_length(label) BETWEEN 1 AND 180),
 description TEXT NOT NULL DEFAULT '' CHECK(char_length(description)<=500),
 kind TEXT NOT NULL CHECK(kind IN ('text','textarea','url','number','checkbox','select','phone')),
 required BOOLEAN NOT NULL DEFAULT false,
 position INTEGER NOT NULL CHECK(position BETWEEN 0 AND 19),
 options TEXT[] NOT NULL DEFAULT '{}' CHECK(cardinality(options)<=20),
 active BOOLEAN NOT NULL DEFAULT true,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_registration_questions_event ON event_registration_questions(event_id,position) WHERE active;

-- Historical submissions survive question edits, cancellation and re-registration.
CREATE TABLE registration_submissions (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 rsvp_id UUID NOT NULL REFERENCES rsvps(id),
 form_revision BIGINT NOT NULL CHECK(form_revision>0),
 answers JSONB NOT NULL CHECK(jsonb_typeof(answers)='array' AND jsonb_array_length(answers)<=20),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_registration_submissions_rsvp ON registration_submissions(rsvp_id,id DESC);
