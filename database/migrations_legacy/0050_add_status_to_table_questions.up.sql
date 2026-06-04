ALTER TABLE public.questions
    ADD COLUMN IF NOT EXISTS status boolean DEFAULT false;

