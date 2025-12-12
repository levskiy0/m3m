import { useState, useRef } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Camera, Key } from 'lucide-react';
import { toast } from 'sonner';

import { usersApi } from '@/api';
import { useAuth } from '@/providers/auth-provider';
import { useTitle } from '@/hooks';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { LoadingButton } from '@/components/ui/loading-button';
import { Field, FieldGroup, FieldLabel, FieldDescription, FieldError } from '@/components/ui/field';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { formatDate } from '@/lib/format';

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

export function ProfilePage() {
  useTitle('Profile');
  const { user, refresh } = useAuth();
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [name, setName] = useState(user?.name || '');
  const [, setNameChanged] = useState(false);

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');

  const updateNameMutation = useMutation({
    mutationFn: () => usersApi.updateMe({ name }),
    onSuccess: () => {
      refresh();
      setNameChanged(false);
      toast.success('Name updated');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update name');
    },
  });

  const updateAvatarMutation = useMutation({
    mutationFn: (file: File) => usersApi.updateAvatar(file),
    onSuccess: () => {
      refresh();
      queryClient.invalidateQueries({ queryKey: ['users'] });
      toast.success('Avatar updated');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to update avatar');
    },
  });

  const changePasswordMutation = useMutation({
    mutationFn: () =>
      usersApi.changePassword({ currentPassword, newPassword }),
    onSuccess: () => {
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      toast.success('Password changed');
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : 'Failed to change password');
    },
  });

  const handleNameChange = (value: string) => {
    setName(value);
    setNameChanged(value !== user?.name);
  };

  const handleAvatarChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      updateAvatarMutation.mutate(file);
    }
    e.target.value = '';
  };

  const handleChangePassword = () => {
    setPasswordError('');

    if (newPassword.length < 8) {
      setPasswordError('Password must be at least 8 characters');
      return;
    }

    if (newPassword !== confirmPassword) {
      setPasswordError('Passwords do not match');
      return;
    }

    changePasswordMutation.mutate();
  };

  if (!user) {
    return null;
  }

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Profile</h1>
        <p className="text-muted-foreground">
          Manage your account settings
        </p>
      </div>

      {/* Avatar */}
      <Card>
        <CardHeader>
          <CardTitle>Avatar</CardTitle>
          <CardDescription>
            Click to upload a new profile picture
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            <div className="relative group">
              <Avatar className="size-24">
                <AvatarImage src={user.avatar} />
                <AvatarFallback className="text-2xl">
                  {getInitials(user.name)}
                </AvatarFallback>
              </Avatar>
              <button
                onClick={() => fileInputRef.current?.click()}
                className="absolute inset-0 flex items-center justify-center bg-black/50 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
                disabled={updateAvatarMutation.isPending}
              >
                <Camera className="size-6 text-white" />
              </button>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                className="hidden"
                onChange={handleAvatarChange}
              />
            </div>
            <div>
              <p className="font-medium">{user.name}</p>
              <p className="text-sm text-muted-foreground">{user.email}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Profile Info */}
      <Card>
        <CardHeader>
          <CardTitle>Profile Information</CardTitle>
          <CardDescription>
            Update your account details
          </CardDescription>
        </CardHeader>
        <CardContent>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="email">Email</FieldLabel>
              <Input id="email" value={user.email} disabled />
              <FieldDescription>
                Email cannot be changed
              </FieldDescription>
            </Field>
            <Field>
              <FieldLabel htmlFor="name">Name</FieldLabel>
              <Input
                id="name"
                value={name}
                onChange={(e) => handleNameChange(e.target.value)}
              />
            </Field>
            <LoadingButton
              className="w-fit"
              onClick={() => updateNameMutation.mutate()}
              loading={updateNameMutation.isPending}
            >
              Save Changes
            </LoadingButton>
          </FieldGroup>
        </CardContent>
      </Card>

      {/* Change Password */}
      <Card>
        <CardHeader>
          <CardTitle>Change Password</CardTitle>
          <CardDescription>
            Update your password to keep your account secure
          </CardDescription>
        </CardHeader>
        <CardContent>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="currentPassword">Current Password</FieldLabel>
              <Input
                id="currentPassword"
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="newPassword">New Password</FieldLabel>
              <Input
                id="newPassword"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
              />
              <FieldDescription>
                Minimum 8 characters
              </FieldDescription>
            </Field>
            <Field>
              <FieldLabel htmlFor="confirmPassword">Confirm New Password</FieldLabel>
              <Input
                id="confirmPassword"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
              />
            </Field>
            {passwordError && (
              <FieldError>{passwordError}</FieldError>
            )}
            <Button
              onClick={handleChangePassword}
              disabled={
                !currentPassword ||
                !newPassword ||
                !confirmPassword ||
                changePasswordMutation.isPending
              }
            >
              <Key className="mr-2 size-4" />
              {changePasswordMutation.isPending
                ? 'Changing...'
                : 'Change Password'}
            </Button>
          </FieldGroup>
        </CardContent>
      </Card>

      {/* Account Info */}
      <Card>
        <CardHeader>
          <CardTitle>Account Information</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Account Type</span>
              <span>{user.is_root ? 'Root Administrator' : 'User'}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Can Create Projects</span>
              <span>{user.permissions.create_projects || user.is_root ? 'Yes' : 'No'}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Can Manage Users</span>
              <span>{user.permissions.manage_users || user.is_root ? 'Yes' : 'No'}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Member Since</span>
              <span>{formatDate(user.created_at)}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
