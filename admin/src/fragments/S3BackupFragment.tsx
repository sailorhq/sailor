import React, { useState } from "react";
import { Badge, Button, Form, Modal } from "react-bootstrap";
import { showToast } from "../utils/toastUtils";
import { useSelector } from "react-redux";
import type { RootState } from "../store";
import { setBucket } from "../backupSlice";
import { useDispatch } from "react-redux";
import { addAuditEvent } from "../audit";
import { useAuth } from "../contexts/AuthContext";

interface S3BackupFragmentProps {
    editMode?: boolean;
}

const S3BackupFragment: React.FC<S3BackupFragmentProps> = (props) => {
    const dispatch = useDispatch();
    const [s3Region, setS3Region] = useState('');
    const [s3AccessKey, setS3AccessKey] = useState('');
    const [s3SecretKey, setS3SecretKey] = useState('');
    const [cronExpression, setCronExpression] = useState('');
    const [savingS3, setSavingS3] = useState(false);

    const [showChangeBucketModal, setShowChangeBucketModal] = useState(false);

    const { bucket } = useSelector((state: RootState) => state.backup);
    const [s3Bucket, setS3Bucket] = useState(bucket || '');

    const { user } = useAuth();


    const handleSaveS3 = async (e: React.FormEvent<any>) => {
        e.preventDefault();

        setSavingS3(true);
        const payload = {
            bucket: s3Bucket,
            access_key: s3AccessKey,
            secret_key: s3SecretKey,
            region: s3Region,
            schedule: cronExpression === '' ? '0 */24 * * *' : cronExpression
        };
        const response = await fetch('/api/v1/admin.backup', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            const data = await response.json();
            showToast(data.message, 3000);
        } else {
            dispatch(setBucket(s3Bucket));
            setS3Region('');
            setS3AccessKey('');
            setS3SecretKey('');
            setCronExpression('');
            showToast('S3 backup details saved successfully', 3000);

            addAuditEvent({
                timestamp: new Date().toISOString(),
                namespace: 'admin',
                app: 'settings',
                username: user?.username || '',
                action: 'update_s3_details',
                details: {
                    bucket: s3Bucket,
                }
            });

            setBucket('');
        }
        setSavingS3(false);
    };

    return (
        <div>
            {bucket && !props.editMode ? (
                <div >
                    <Badge bg="success">Bucket: {bucket}</Badge>
                    <br />
                    <br />
                    <Button onClick={() => setShowChangeBucketModal(true)} style={{ backgroundColor: "#1C608C", border: "none" }}>Change</Button>
                </div>
            ) : <Form onSubmit={handleSaveS3}>
                <Form.Group className="mb-3" controlId="formS3Region">
                    <Form.Label>Bucket</Form.Label>
                    <Form.Control
                        required
                        type="text"
                        placeholder="Enter Bucket Name"
                        value={s3Bucket}
                        onChange={e => setS3Bucket(e.target.value)}
                    />
                </Form.Group>

                <Form.Group className="mb-3" controlId="formS3Region">
                    <Form.Label>S3 Region</Form.Label>
                    <Form.Control
                        required
                        type="text"
                        placeholder="Enter S3 Region (eg: us-east-1)"
                        value={s3Region}
                        onChange={e => setS3Region(e.target.value)}
                    />
                </Form.Group>

                <Form.Group className="mb-3" controlId="formS3AccessKey">
                    <Form.Label>S3 Access Key</Form.Label>
                    <Form.Control
                        required
                        type="password"
                        placeholder="Enter S3 Access Key"
                        value={s3AccessKey}
                        onChange={e => setS3AccessKey(e.target.value)}
                    />
                </Form.Group>

                <Form.Group className="mb-3" controlId="formS3SecretKey">
                    <Form.Label>S3 Secret Key</Form.Label>
                    <Form.Control
                        required
                        type="password"
                        placeholder="Enter S3 Secret Key"
                        value={s3SecretKey}
                        onChange={e => setS3SecretKey(e.target.value)}
                    />
                </Form.Group>

                <Form.Group className="mb-3" controlId="formS3CronSchedule">
                    <Form.Label style={{ color: '#1C608C' }}>Schedule (Cron)</Form.Label>
                    <Form.Control
                        type="text"
                        placeholder="Enter cron expression"
                        value={cronExpression}
                        onChange={e => setCronExpression(e.target.value)}
                    />
                </Form.Group>
                <div className="d-flex justify-content-end gap-2">
                    <Button type="submit" style={{ backgroundColor: '#1C608C', border: 'none' }} disabled={savingS3}>
                        {savingS3 ? 'Saving...' : 'Save'}
                    </Button>
                </div>
            </Form>}

            <Modal style={{ fontFamily: 'Quicksand' }} show={showChangeBucketModal} onHide={() => setShowChangeBucketModal(false)}>
                <Modal.Header closeButton>
                    <Modal.Title>Change Bucket</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <S3BackupFragment editMode={true} />
                </Modal.Body>
            </Modal>
        </div>
    );
};

export default S3BackupFragment;