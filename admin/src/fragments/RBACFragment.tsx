import { useState } from "react";
import { Button, Form } from "react-bootstrap";
import type { RootState } from "../store";
import { useSelector } from "react-redux";
import type { User } from "../contexts/AuthContext";
import { showToast } from "../utils/toastUtils";

type RBACFragmentProps = {
    user: User | null,
    onUserCreated: (() => void) | null
}

const RBACFragment: React.FC<RBACFragmentProps> = (props: RBACFragmentProps) => {
    const [usernameToCreate, setUsernameToCreate] = useState(props.user?.username || '');
    const [passwordToCreate, setPasswordToCreate] = useState('');
    const [roleToCreate, setRoleToCreate] = useState(props.user?.roles[0] || '');
    const [permissionsToCreate, setPermissionsToCreate] = useState<Set<string>>(new Set(props.user?.permissions || []));
    const [allowedApps, setAllowedApps] = useState<Set<string>>(new Set(props.user?.allowed_apps || []));
    const [savingS3] = useState(false);


    const apps = useSelector((state: RootState) => state.apps.values);

    const handleUserCreation = async () => {
        // Implementation of handleUserCreation
        const permissions = Array.from(permissionsToCreate).join('|');
        const joinedAllowedApps = Array.from(allowedApps).join('|');
        const api = props.user !== null ? 'admin.edit.user' : 'admin.create.user';
        const url = `http://localhost:7766/api/v1/${api}?username=${usernameToCreate}&password=${passwordToCreate}&role=${roleToCreate}&permissions=${permissions}&allowed_apps=${joinedAllowedApps}`;
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (response.ok) {
            setUsernameToCreate('');
            setPasswordToCreate('');
            setRoleToCreate('');
            setPermissionsToCreate(new Set());
            setAllowedApps(new Set());

            if (props.onUserCreated) {
                props.onUserCreated();
            }

            showToast(`User ${props.user !== null ? 'updated' : 'created'} successfully`, 3000);
        }
    };

    const handlePermissionChange = (permission: string, checked: boolean) => {
        if (checked) {
            setPermissionsToCreate(prev => new Set([...prev, permission]))
        } else {
            const updated = new Set([...permissionsToCreate])
            updated.delete(permission);
            setPermissionsToCreate(updated);
        }
    }

    return (
        <div>
            <Form>
                {/* <Form.Check
                                    type="radio"
                                    label="Light"
                                    name="themeRadios"
                                    id="theme-light"
                                    value="option1"
                                    checked={radioValue === 'option1'}
                                    onChange={() => setRadioValue('option1')}
                                />
                                <Form.Check
                                    type="radio"
                                    label="Dark"
                                    name="themeRadios"
                                    id="theme-dark"
                                    value="option2"
                                    checked={radioValue === 'option2'}
                                    onChange={() => setRadioValue('option2')}
                                />
                                <hr className="my-4" /> */}
                <Form.Group className="mb-3" controlId="formS3AccessKey">
                    <Form.Label>Username</Form.Label>
                    <Form.Control
                        disabled={props.user !== null}
                        required
                        type="text"
                        placeholder="Enter username"
                        value={usernameToCreate}
                        onChange={e => setUsernameToCreate(e.target.value)}
                    />
                </Form.Group>

                {props.user === null && <Form.Group className="mb-3" controlId="formS3AccessKey">
                    <Form.Label>Password</Form.Label>
                    <Form.Control
                        required
                        type="password"
                        placeholder="Enter password"
                        value={passwordToCreate}
                        onChange={e => setPasswordToCreate(e.target.value)}
                    />
                </Form.Group>}
                <Form.Group className="mb-3" controlId="formS3AccessKey">
                    <Form.Label>Role</Form.Label>
                    <Form.Select
                        required
                        value={roleToCreate}
                        onChange={e => setRoleToCreate(e.target.value)}
                    >
                        <option value="">Select role</option>
                        <option value="admin">Admin</option>
                        <option value="user">User</option>
                        <option value="viewer">Viewer</option>
                    </Form.Select>
                </Form.Group>
                <Form.Group className="mb-3" controlId="formS3SecretKey">
                    <Form.Label>Permissions</Form.Label>
                    <div style={{ width: 'inherit' }} className="d-flex flex-column gap-2">
                        <Form.Check
                            checked={permissionsToCreate.has('create_configs')}
                            type="checkbox"
                            id="perm-create-configs"
                            label="create_configs"
                            onChange={e => handlePermissionChange('create_configs', e.target.checked)}
                        />
                        <Form.Check
                            checked={permissionsToCreate.has('edit_schema')}
                            type="checkbox"
                            id="perm-update-pod-url"
                            label="edit_schema"
                            onChange={e => handlePermissionChange('edit_schema', e.target.checked)}
                        />
                        <Form.Check
                            checked={permissionsToCreate.has('deploy_configs')}
                            type="checkbox"
                            id="perm-deploy-configs"
                            label="deploy_configs"
                            onChange={e => handlePermissionChange('deploy_configs', e.target.checked)}
                        />
                        <Form.Check
                            checked={permissionsToCreate.has('create_secrets')}
                            type="checkbox"
                            id="perm-create-secrets"
                            label="create_secrets"
                            onChange={e => handlePermissionChange('create_secrets', e.target.checked)}
                        />
                        <Form.Check
                            checked={permissionsToCreate.has('edit_policy')}
                            type="checkbox"
                            id="perm-update-pod-url"
                            label="edit_policy"
                            onChange={e => handlePermissionChange('edit_policy', e.target.checked)}
                        />
                        <Form.Check
                            checked={permissionsToCreate.has('deploy_secrets')}
                            type="checkbox"
                            id="perm-deploy-secrets"
                            label="deploy_secrets"
                            onChange={e => handlePermissionChange('deploy_secrets', e.target.checked)}
                        />
                        <Form.Check
                            checked={permissionsToCreate.has('update_pod_url')}
                            type="checkbox"
                            id="perm-update-pod-url"
                            label="update_pod_url"
                            onChange={e => handlePermissionChange('update_pod_url', e.target.checked)}
                        />


                    </div>
                </Form.Group>

                <Form.Group className="mb-3" controlId="formS3SecretKey">
                    <Form.Label>Allowed Apps</Form.Label>
                    <div className="d-flex flex-column gap-2">
                        {Object.values(apps).flat().map(app => <Form.Check
                            checked={allowedApps.has(app)}
                            type="checkbox"
                            id={app}
                            label={app}
                            onChange={(e) => {
                                if (e.target.checked) {
                                    setAllowedApps(prev => new Set([...prev, app]))
                                } else {
                                    const updated = new Set([...allowedApps])
                                    updated.delete(app);
                                    setAllowedApps(updated);
                                }
                            }}
                        />)}
                    </div>
                </Form.Group>
                <div className="d-flex gap-2">
                    <Button style={{ backgroundColor: '#1C608C', border: 'none' }} onClick={handleUserCreation} disabled={savingS3}>
                        {savingS3 ? 'Saving...' : 'Save'}
                    </Button>
                </div>
            </Form>
        </div>
    );
};

export default RBACFragment;