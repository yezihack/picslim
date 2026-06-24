// Wails bindings - use generated wailsjs
import * as AppBinding from '../../wailsjs/go/app/App';
import * as Runtime from '../../wailsjs/runtime/runtime';

import {
  BasicResult,
  SelectPathsResult,
  SelectDirResult,
  ScanResult,
  ExportResult,
  TaskStatusResult,
  CompletedJobResult,
  PreviewResult,
} from '../types';

// Check if running in Wails
export function isWails(): boolean {
  return typeof window !== 'undefined' &&
         typeof (window as any).go !== 'undefined' &&
         typeof (window as any).go.app !== 'undefined';
}

// API service - directly use Wails bindings
export const api = {
  // Select input paths
  async selectInputPaths(): Promise<SelectPathsResult> {
    return AppBinding.SelectInputPaths();
  },

  // Select input directory
  async selectInputDir(): Promise<SelectDirResult> {
    return AppBinding.SelectInputDir();
  },

  // Select output directory
  async selectOutputDir(): Promise<SelectDirResult> {
    return AppBinding.SelectOutputDir();
  },

  // Scan input paths
  async scanInputPaths(paths: string[]): Promise<ScanResult> {
    return AppBinding.ScanInputPaths(paths);
  },

  // Start compress
  async startCompress(inputDir: string, outputDir: string, preset: string, concurrency: number): Promise<BasicResult> {
    return AppBinding.StartCompress(inputDir, outputDir, preset, concurrency);
  },

  // Pause task
  async pauseTask(): Promise<BasicResult> {
    return AppBinding.PauseTask();
  },

  // Resume task
  async resumeTask(): Promise<BasicResult> {
    return AppBinding.ResumeTask();
  },

  // Cancel task
  async cancelTask(): Promise<BasicResult> {
    return AppBinding.CancelTask();
  },

  // Get task status
  async getTaskStatus(): Promise<TaskStatusResult> {
    return AppBinding.GetTaskStatus();
  },

  // Open output directory
  async openOutputDir(): Promise<BasicResult> {
    return AppBinding.OpenOutputDir();
  },

  // Export report
  async exportReport(format: string): Promise<ExportResult> {
    return AppBinding.ExportReport(format);
  },

  // Get output directory
  async getOutputDir(): Promise<string> {
    return AppBinding.GetOutputDir();
  },

  // Get completed jobs
  async getCompletedJobs(): Promise<CompletedJobResult> {
    return AppBinding.GetCompletedJobs();
  },

  // Get preview
  async getPreview(index: number): Promise<PreviewResult> {
    return AppBinding.GetPreview(index);
  },
};

// Events
export const events = {
  // Subscribe to init events
  onInit(callback: (data: any) => void): void {
    Runtime.EventsOn('task:init', callback);
  },

  // Subscribe to log events
  onLog(callback: (data: any) => void): void {
    Runtime.EventsOn('task:log', callback);
  },

  // Subscribe to progress events
  onProgress(callback: (data: any) => void): void {
    Runtime.EventsOn('task:progress', callback);
  },

  // Subscribe to completed events
  onCompleted(callback: (data: any) => void): void {
    Runtime.EventsOn('task:completed', callback);
  },

  // Subscribe to cancelled events
  onCancelled(callback: (data: any) => void): void {
    Runtime.EventsOn('task:cancelled', callback);
  },

  // Subscribe to paused events
  onPaused(callback: (data: any) => void): void {
    Runtime.EventsOn('task:paused', callback);
  },

  // Subscribe to resumed events
  onResumed(callback: (data: any) => void): void {
    Runtime.EventsOn('task:resumed', callback);
  },

  // Unsubscribe from all events
  offAll(): void {
    Runtime.EventsOff('task:init');
    Runtime.EventsOff('task:log');
    Runtime.EventsOff('task:progress');
    Runtime.EventsOff('task:completed');
    Runtime.EventsOff('task:cancelled');
    Runtime.EventsOff('task:paused');
    Runtime.EventsOff('task:resumed');
  },
};