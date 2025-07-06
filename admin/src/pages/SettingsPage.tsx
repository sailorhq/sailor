import React, { useState } from 'react';
import { Card, Accordion, Form, Button, Modal, ListGroup } from 'react-bootstrap';
import { useAuth, type User } from '../contexts/AuthContext';
import RBACFragment from '../fragments/RBACFragment';
import S3BackupFragment from '../fragments/S3BackupFragment';
import { showToast } from '../utils/toastUtils';

const SettingsPage: React.FC = () => {
    const [showChangePasswordModal, setShowChangePasswordModal] = useState(false);
    const [showUsersModal, setShowUsersModal] = useState(false);
    const [pwd, setPwd] = useState('');
    const [switchOn, setSwitchOn] = useState(false);

    const [selectedUser, setSelectedUser] = useState<User | null>(null);
    const [showEditUserModal, setShowEditUserModal] = useState(false);
    const [loading, setLoading] = useState(false);

    // Sample users data - in a real app, this would come from an API
    const [users, setUsers] = useState<User[]>([]);

    const { user, logout } = useAuth();



    const handleChangePassword = async (e: React.FormEvent<any>) => {
        e.preventDefault();
        setLoading(true);

        const response = await fetch('/api/v1/admin.change.password', {
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
        setLoading(false);
    };

    // const handleRevokeUser = (userId: number) => {
    // Remove user from the list
    // setUsers(users.filter(user => user.id !== userId));
    // In a real app, you would make an API call here to revoke the user
    // };


    const fetchUsers = async () => {
        const response = await fetch('/api/v1/admin.list.users');
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
                            <S3BackupFragment />
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