import React, { useState } from 'react';
import { Card, Accordion, Form, Button, Modal, ListGroup } from 'react-bootstrap';
import { useAuth, type User } from '../contexts/AuthContext';
import RBACFragment from '../fragments/RBACFragment';
import { showToast } from '../utils/toastUtils';

const SettingsPage: React.FC = () => {
    const [showModal, setShowModal] = useState(false);
    const [showChangePasswordModal, setShowChangePasswordModal] = useState(false);
    const [showUsersModal, setShowUsersModal] = useState(false);
    const [selectedSetting, setSelectedSetting] = useState('');
    const [pwd, setPwd] = useState('');
    const [switchOn, setSwitchOn] = useState(false);
    const [bucketHost, setBucketHost] = useState('');
    const [s3AccessKey, setS3AccessKey] = useState('');
    const [s3SecretKey, setS3SecretKey] = useState('');
    const [checkingS3, setCheckingS3] = useState(false);
    const [savingS3, setSavingS3] = useState(false);
    const [loading, setLoading] = useState(false);
    const [selectedUser, setSelectedUser] = useState<User | null>(null);
    const [showEditUserModal, setShowEditUserModal] = useState(false);

    // Sample users data - in a real app, this would come from an API
    const [users, setUsers] = useState<User[]>([]);

    const { user, logout } = useAuth();


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

    const handleChangePassword = async (e: React.FormEvent<any>) => {
        e.preventDefault();


        const response = await fetch('http://localhost:7766/api/v1/admin.change.password', {
            method: 'POST',
            headers: {
                'x-username': user?.username || '',
                'x-password': pwd
            }
        });

        if (!response.ok) {
            const data = await response.json();
            showToast(data.message, 3000);
            return;
        }

        logout();
        showToast('Password changed successfully', 3000);
        setShowChangePasswordModal(false);
    };

    const handleRevokeUser = (userId: number) => {
        // Remove user from the list
        // setUsers(users.filter(user => user.id !== userId));
        // In a real app, you would make an API call here to revoke the user
    };


    const fetchUsers = async () => {
        const response = await fetch('http://localhost:7766/api/v1/admin.list.users');
        const data = await response.json();
        setUsers(data);
    }

    const handleOpenUsersModal = async () => {
        await fetchUsers();
        setShowUsersModal(true);
    }

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
                                    <Button style={{ border: '1px solid#1C608C', backgroundColor: '#FEFBF0', color: 'black' }} onClick={handleCheckS3} disabled={checkingS3}>
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
                        <Accordion.Header>Security</Accordion.Header>
                        <Accordion.Body>
                            <ListGroup variant="flush">
                                <ListGroup.Item action onClick={() => setShowChangePasswordModal(true)}>
                                    Change Default Password
                                </ListGroup.Item>
                            </ListGroup>
                        </Accordion.Body>
                    </Accordion.Item>

                    <Accordion.Item eventKey="4">
                        <Accordion.Header>Users & Roles</Accordion.Header>
                        <Accordion.Body>
                            <div className="d-flex justify-content-end">
                                <Button style={{ backgroundColor: '#1C608C', border: 'none' }} onClick={handleOpenUsersModal}>
                                    View/Edit Users & Roles
                                </Button>
                            </div>
                            <RBACFragment user={null} onUserCreated={null} />
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
                                    type="password"
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
                                <ListGroup.Item onClick={() => {
                                    setSelectedUser(user);
                                    setShowEditUserModal(true);
                                }} style={{ cursor: 'pointer' }} key={user.id} className="d-flex justify-content-between align-items-center">
                                    <div className="flex-grow-1">
                                        <div className="d-flex justify-content-between align-items-start mb-2">
                                            <h6 className="mb-0">{user.username}</h6>
                                            <div className="text-muted small">
                                                {user.roles.join(', ')}
                                            </div>
                                        </div>
                                    </div>
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

                <Modal backdrop="static" style={{ fontFamily: 'Quicksand' }} show={showEditUserModal} onHide={() => setShowEditUserModal(false)} centered size="lg">
                    <Modal.Header closeButton>
                        <Modal.Title>User Details</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <RBACFragment user={selectedUser} onUserCreated={() => {
                            fetchUsers();
                            setShowEditUserModal(false);
                            setSelectedUser(null);
                        }} />
                    </Modal.Body>
                </Modal>
            </Card.Body>
        </Card>
    );
};

export default SettingsPage; 