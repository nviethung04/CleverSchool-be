-- Migration: Drop homework_questions, exam_questions, exercise_questions tables
-- Date: 2025-01-08
-- Description: Drop the three questions tables if they exist

-- Drop homework_questions table if exists
DROP TABLE IF EXISTS homework_questions;

-- Drop exam_questions table if exists  
DROP TABLE IF EXISTS exam_questions;

-- Drop exercise_questions table if exists
DROP TABLE IF EXISTS exercise_questions;
