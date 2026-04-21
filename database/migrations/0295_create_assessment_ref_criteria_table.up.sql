-- Table: public.assessment_ref_criteria

-- DROP TABLE IF EXISTS public.assessment_ref_criteria;

CREATE TABLE IF NOT EXISTS public.assessment_ref_criteria
(
    id bigserial,
    assessment_id bigint,
    assessment_criterion_id bigint,
    created_at timestamp without time zone,
    created_by bigint,
    updated_at timestamp without time zone,
    updated_by bigint,
    deleted_at timestamp without time zone,
    deleted_by bigint,
    CONSTRAINT assessment_ref_criteria_pkey PRIMARY KEY (id)
)
TABLESPACE pg_default;

-- Index: assessment_ref_criteria_assessment_criterion_id_idx

-- DROP INDEX IF EXISTS public.assessment_ref_criteria_assessment_criterion_id_idx;

CREATE INDEX IF NOT EXISTS assessment_ref_criteria_assessment_criterion_id_idx
    ON public.assessment_ref_criteria USING btree
    (assessment_criterion_id ASC NULLS LAST)
    TABLESPACE pg_default;

-- Index: assessment_ref_criteria_assessment_id_idx

-- DROP INDEX IF EXISTS public.assessment_ref_criteria_assessment_id_idx;

CREATE INDEX IF NOT EXISTS assessment_ref_criteria_assessment_id_idx
    ON public.assessment_ref_criteria USING btree
    (assessment_id ASC NULLS LAST)
    TABLESPACE pg_default;

