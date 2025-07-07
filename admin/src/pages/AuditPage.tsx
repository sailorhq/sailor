import React, { useEffect, useState } from 'react';
import { Card, ListGroup, Badge, Button } from 'react-bootstrap';
import { getAuditEvents, type AuditEvent } from '../audit';

const getBadgeClassName = (action: string) => {
    switch (action) {
        case 'create_app':
        case 'create_deployment':
            return 'badge-create-app';
        case 'update_schema':
        case 'update_user':
            return 'badge-update';
        default:
            return 'badge-deploy-configs';
    }
}

const AuditPage: React.FC = () => {
    const [auditEvents, setAuditEvents] = useState<AuditEvent[]>([]);

    useEffect(() => {
        getAuditEvents().then(setAuditEvents);
    }, []);

    return (
        <Card className="border-0 shadow-sm" style={{ minHeight: '80vh', borderRadius: 10 }}>
            <Card.Body>
                <Card.Title className="mb-2">Audit Trail</Card.Title>
                <div className="d-flex justify-content-end">
                    <Button className='mb-4' style={{ backgroundColor: '#1C608C', border: 'none' }} onClick={() => getAuditEvents().then(setAuditEvents)}>Refresh</Button>
                </div>
                <ListGroup style={{ maxHeight: '70vh', overflowY: 'auto' }}>
                    {auditEvents.map((event, index) => (
                        <ListGroup.Item key={index} className="d-flex justify-content-between align-items-start">
                            <div className="ms-2 me-auto">
                                <div className="text-muted">
                                    <Badge className={`me-2 ${getBadgeClassName(event.action)}`}>{event.action}</Badge>
                                    {JSON.stringify(event, null, 2)}
                                </div>
                            </div>
                        </ListGroup.Item>
                    ))}
                    {auditEvents.length === 0 && (
                        <ListGroup.Item className="text-center text-muted">
                            No audit events found
                        </ListGroup.Item>
                    )}
                </ListGroup>
            </Card.Body>
        </Card>
    );
};

export default AuditPage; 