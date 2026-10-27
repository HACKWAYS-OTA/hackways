import type { Event, Space } from './index';

export type WorkspaceRole = 'owner' | 'manager' | 'editor' | 'checkin';
export type WorkspacePermission = 'view' | 'content' | 'lifecycle' | 'guests.read' | 'guests.manage' | 'guests.export' | 'checkin' | 'create' | 'access.manage' | 'delete' | 'association';
export interface WorkspaceAccess { resource_id: string; kind: 'space' | 'event'; role: WorkspaceRole; permissions: WorkspacePermission[] }
export type WorkspaceEvent = Event & { access_role: WorkspaceRole };
export type WorkspaceSpace = Pick<Space, 'id' | 'slug' | 'name' | 'description' | 'logo_url' | 'banner_url' | 'status' | 'revision' | 'updated_at'> & { access_role: WorkspaceRole };
export interface WorkspaceMember { id: string; user_id: string; display_name: string; email: string; role: Exclude<WorkspaceRole, 'owner'>; revision: number; revoked_at: string | null }
export interface WorkspaceTeam { data: WorkspaceMember[]; inherited: WorkspaceMember[]; limit: number; offset: number }
