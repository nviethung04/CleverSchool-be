-- Drop indexes
DROP INDEX IF EXISTS public.assessment_ref_criteria_assessment_id_idx;
DROP INDEX IF EXISTS public.assessment_ref_criteria_assessment_criterion_id_idx;

-- Drop table
DROP TABLE IF EXISTS public.assessment_ref_criteria;

