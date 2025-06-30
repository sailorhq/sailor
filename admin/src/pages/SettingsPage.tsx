import React, { useState } from 'react';
import { Card, Accordion, Form, Button, Modal, ListGroup } from 'react-bootstrap';

const SettingsPage: React.FC = () => {
    const [showModal, setShowModal] = useState(false);
    const [showChangePasswordModal, setShowChangePasswordModal] = useState(false);
    const [showUsersModal, setShowUsersModal] = useState(false);
    const [selectedSetting, setSelectedSetting] = useState('');
    const [pwd, setPwd] = useState('');
    const [radioValue, setRadioValue] = useState('option1');
    const [switchOn, setSwitchOn] = useState(false);
    const [bucketHost, setBucketHost] = useState('');
    const [s3AccessKey, setS3AccessKey] = useState('');
    const [s3SecretKey, setS3SecretKey] = useState('');
    const [checkingS3, setCheckingS3] = useState(false);
    const [savingS3, setSavingS3] = useState(false);
    const [loading, setLoading] = useState(false);
    const [usernameToCreate, setUsernameToCreate] = useState('');
    const [passwordToCreate, setPasswordToCreate] = useState('');
    const [roleToCreate, setRoleToCreate] = useState('');
    const [permissionsToCreate, setPermissionsToCreate] = useState<Set<string>>(new Set());

    const apps = useSelector((state: RootState) => state.apps.values);

    // Sample users data - in a real app, this would come from an API
    const [users, setUsers] = useState([
        { id: 1, username: 'admin', role: 'Admin', permissions: ['read', 'write', 'delete', 'admin'] },
        { id: 2, username: 'john_doe', role: 'User', permissions: ['read', 'write'] },
        { id: 3, username: 'jane_smith', role: 'Viewer', permissions: ['read'] },
        { id: 4, username: 'bob_wilson', role: 'User', permissions: ['read', 'write'] },
    ]);

    const handleOpenModal = (setting: string) => {
        setSelectedSetting(setting);
        setShowModal(true);
    };

    const handleCloseModal = () => {
        setShowModal(false);
        setSelectedSetting('');
    };

    const handleCheckS3 = () => {
        setCheckingS3(true);
        // Implementation of handleCheckS3
    };

    const handleSaveS3 = () => {
        setSavingS3(true);
        // Implementation of handleSaveS3
    };

    const handleChangePassword = () => {
        setLoading(true);
        // Implementation of handleChangePassword
        setLoading(false);
        setShowChangePasswordModal(true);
    };

    const handleRevokeUser = (userId: number) => {
        // Remove user from the list
        setUsers(users.filter(user => user.id !== userId));
        // In a real app, you would make an API call here to revoke the user
    };

    const handleUserCreation = async () => {
        // Implementation of handleUserCreation
        const permissions = Array.from(permissionsToCreate).join('|');
        const url = `http://localhost:7766/api/v1/admin.create.user?username=${usernameToCreate}&password=${passwordToCreate}&role=${roleToCreate}&permissions=${permissions}`;
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (response.status != 200) {
            alert('Failed to create user');
        }
    };

    return (
        <Card className="border-0 shadow-sm" style={{ minHeight: '80vh', borderRadius: 10 }}>
            <Card.Body>
                <h4 className="mb-4">Settings</h4>
                <Accordion defaultActiveKey="0">
                    <Accordion.Item eventKey="0">
                        <Accordion.Header>S3 Backup & Fallback</Accordion.Header>
                        <Accordion.Body>
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
                                    <Form.Label>S3 Host</Form.Label>
                                    <Form.Control
                                        type="text"
                                        placeholder="Enter S3 Host"
                                        value={bucketHost}
                                        onChange={e => setBucketHost(e.target.value)}
                                    />
                                </Form.Group>
                                <Form.Group className="mb-3" controlId="formS3AccessKey">
                                    <Form.Label>S3 Access Key</Form.Label>
                                    <Form.Control
                                        type="text"
                                        placeholder="Enter S3 Access Key"
                                        value={s3AccessKey}
                                        onChange={e => setS3AccessKey(e.target.value)}
                                    />
                                </Form.Group>
                                <Form.Group className="mb-3" controlId="formS3SecretKey">
                                    <Form.Label>S3 Secret Key</Form.Label>
                                    <Form.Control
                                        type="password"
                                        placeholder="Enter S3 Secret Key"
                                        value={s3SecretKey}
                                        onChange={e => setS3SecretKey(e.target.value)}
                                    />
                                </Form.Group>
                                <div className="d-flex gap-2">
                                    <Button style={{ border: '1px solid#1C608C', backgroundColor: '#FEF8EE', color: 'black' }} onClick={handleCheckS3} disabled={checkingS3}>
                                        {checkingS3 ? 'Checking...' : 'Check'}
                                    </Button>
                                    <Button style={{ backgroundColor: '#1C608C', border: 'none' }} onClick={handleSaveS3} disabled={savingS3}>
                                        {savingS3 ? 'Saving...' : 'Save'}
                                    </Button>
                                </div>
                            </Form>
                        </Accordion.Body>
                    </Accordion.Item>
                    <Accordion.Item eventKey="1">
                        <Accordion.Header>Logging</Accordion.Header>
                        <Accordion.Body>
                            <Form>
                                <Form.Check
                                    type="switch"
                                    id="custom-switch"
                                    label="Enable Deployment Logging"
                                    checked={switchOn}
                                    onChange={() => setSwitchOn(!switchOn)}
                                />
                            </Form>
                        </Accordion.Body>
                    </Accordion.Item>
                    <Accordion.Item eventKey="2">
                        <Accordion.Header>Panel</Accordion.Header>
                        <Accordion.Body>
                            <ListGroup variant="flush">
                                <ListGroup.Item action onClick={() => handleChangePassword()}>
                                    Change Default Password
                                </ListGroup.Item>
                                {/* <ListGroup.Item action onClick={() => handleOpenModal('Advanced Setting 2')}>
                                    Advanced Setting 2
                                </ListGroup.Item> */}
                            </ListGroup>
                        </Accordion.Body>
                    </Accordion.Item>

                    <Accordion.Item eventKey="4">
                        <Accordion.Header>Users & Roles</Accordion.Header>
                        <Accordion.Body>
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
                                        type="text"
                                        placeholder="Enter username"
                                        value={usernameToCreate}
                                        onChange={e => setUsernameToCreate(e.target.value)}
                                    />
                                </Form.Group>

                                <Form.Group className="mb-3" controlId="formS3AccessKey">
                                    <Form.Label>Password</Form.Label>
                                    <Form.Control
                                        type="text"
                                        placeholder="Enter password"
                                        value={passwordToCreate}
                                        onChange={e => setPasswordToCreate(e.target.value)}
                                    />
                                </Form.Group>
                                <Form.Group className="mb-3" controlId="formS3AccessKey">
                                    <Form.Label>Role</Form.Label>
                                    <Form.Select
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
                                    <div className="d-flex flex-row gap-4">
                                        <Form.Check
                                            type="checkbox"
                                            id="perm-create-configs"
                                            label="create_configs"
                                            onChange={() => setPermissionsToCreate(prev => new Set([...prev, 'create_configs']))}
                                        />
                                        <Form.Check
                                            type="checkbox"
                                            id="perm-create-secrets"
                                            label="create_secrets"
                                            onChange={() => setPermissionsToCreate(prev => new Set([...prev, 'create_secrets']))}
                                        />
                                        <Form.Check
                                            type="checkbox"
                                            id="perm-deploy-configs"
                                            label="deploy_configs"
                                            onChange={() => setPermissionsToCreate(prev => new Set([...prev, 'deploy_configs']))}
                                        />
                                        <Form.Check
                                            type="checkbox"
                                            id="perm-deploy-secrets"
                                            label="deploy_secrets"
                                            onChange={() => setPermissionsToCreate(prev => new Set([...prev, 'deploy_secrets']))}
                                        />
                                        <Form.Check
                                            type="checkbox"
                                            id="perm-update-pod-url"
                                            label="update_pod_url"
                                            onChange={() => setPermissionsToCreate(prev => new Set([...prev, 'update_pod_url']))}
                                        />
                                    </div>
                                </Form.Group>

                                <Form.Group className="mb-3" controlId="formS3SecretKey">
                                    <Form.Label>Allowed Apps</Form.Label>
                                    <div className="d-flex flex-row gap-4">
                                        {Object.values(apps).flat().map(app => <Form.Check
                                            checked={allowedApps.has(app)}
                                            type="checkbox"
                                            id={app}
                                            label={app}
                                            onChange={() => setAllowedApps(prev => new Set([...prev, app]))}
                                        />)}
                                    </div>
                                </Form.Group>
                                <div className="d-flex gap-2">
                                    <Button style={{ backgroundColor: '#1C608C', border: 'none' }} onClick={handleUserCreation} disabled={savingS3}>
                                        {savingS3 ? 'Saving...' : 'Save'}
                                    </Button>

                                    <Button style={{ backgroundColor: '#1C608C', border: 'none' }} onClick={() => setShowUsersModal(true)}>
                                        View/Edit Roles & Permissions
                                    </Button>
                                </div>
                            </Form>
                        </Accordion.Body>
                    </Accordion.Item>
                </Accordion>
                <Modal show={showModal} onHide={handleCloseModal} centered>
                    <Modal.Header closeButton>
                        <Modal.Title>{selectedSetting}</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <p>Settings for <b>{selectedSetting}</b> go here.</p>
                        <Button variant="primary" onClick={handleCloseModal}>Close</Button>
                    </Modal.Body>
                </Modal>

                <Modal style={{ fontFamily: 'Quicksand' }} show={showChangePasswordModal} onHide={() => setShowChangePasswordModal(false)} centered>
                    <Modal.Header closeButton>
                        <Modal.Title>Change Default Password</Modal.Title>
                    </Modal.Header>
                    <Form onSubmit={handleChangePassword}>
                        <Modal.Body>
                            <Form.Group className="mb-3" controlId="formNamespace">
                                <Form.Label>Password</Form.Label>
                                <Form.Control
                                    type="text"
                                    value={pwd}
                                    onChange={e => setPwd(e.target.value)}
                                    required
                                    placeholder="Enter password"
                                />
                            </Form.Group>
                        </Modal.Body>
                        <Modal.Footer>
                            <Button type="submit" style={{ backgroundColor: '#1C608C', border: 'none' }} disabled={loading}>
                                {loading ? 'Creating...' : 'Create'}
                            </Button>
                        </Modal.Footer>
                    </Form>
                </Modal>

                <Modal style={{ fontFamily: 'Quicksand' }} show={showUsersModal} onHide={() => setShowUsersModal(false)} centered size="lg">
                    <Modal.Header closeButton>
                        <Modal.Title>Roles & Permissions</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <ListGroup>
                            {users.map((user) => (
                                <ListGroup.Item key={user.id} className="d-flex justify-content-between align-items-center">
                                    <div className="flex-grow-1">
                                        <div className="d-flex justify-content-between align-items-start mb-2">
                                            <h6 className="mb-0">{user.username}</h6>
                                            <span className={`badge ${user.role === 'Admin' ? 'bg-danger' : user.role === 'User' ? 'bg-primary' : 'bg-secondary'}`}>
                                                {user.role}
                                            </span>
                                        </div>
                                        <div className="text-muted small">
                                            <strong>Permissions:</strong> {user.permissions.join(', ')}
                                        </div>
                                    </div>
                                    <Button
                                        variant="outline-danger"
                                        size="sm"
                                        onClick={() => handleRevokeUser(user.id)}
                                        style={{ marginLeft: '1rem' }}
                                    >
                                        Revoke
                                    </Button>
                                </ListGroup.Item>
                            ))}
                        </ListGroup>
                        {users.length === 0 && (
                            <div className="text-center text-muted py-4">
                                No users found
                            </div>
                        )}
                    </Modal.Body>
                    <Modal.Footer>
                        <Button variant="secondary" onClick={() => setShowUsersModal(false)}>
                            Close
                        </Button>
                    </Modal.Footer>
                </Modal>
            </Card.Body>
        </Card>
    );
};

export default SettingsPage; 