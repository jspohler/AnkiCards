import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Box,
    Typography,
    CircularProgress,
    Paper,
    Alert,
} from '@mui/material';

interface ProcessingStatus {
    status: 'processing' | 'completed' | 'failed';
    progress: number;
    message: string;
    deckName?: string;
    error?: string;
}

export default function ProcessingStatus() {
    const { jobId } = useParams<{ jobId: string }>();
    const navigate = useNavigate();
    const [status, setStatus] = useState<ProcessingStatus>({
        status: 'processing',
        progress: 0,
        message: 'Starting processing...',
    });

    useEffect(() => {
        const checkStatus = async () => {
            try {
                const response = await fetch(`/api/process/${jobId}`);
                const data = await response.json();
                
                if (data.error) {
                    setStatus({
                        status: 'failed',
                        progress: 0,
                        message: data.error,
                        error: data.error
                    });
                    return;
                }

                setStatus(data);

                if (data.status === 'completed') {
                    // Extract filename without extension to use as deck name
                    const deckName = data.deckName || data.filename?.replace(/\.[^/.]+$/, '') || 'default';
                    
                    // Short delay to show completion message
                    setTimeout(() => {
                        navigate(`/review/${deckName}`);
                    }, 1500);
                } else if (data.status === 'processing') {
                    // Continue polling
                    setTimeout(checkStatus, 2000);
                }
            } catch (error) {
                console.error('Error checking status:', error);
                setStatus({
                    status: 'failed',
                    progress: 0,
                    message: 'Failed to check processing status',
                    error: error instanceof Error ? error.message : 'Unknown error'
                });
            }
        };

        checkStatus();
    }, [jobId, navigate]);

    return (
        <Box sx={{ maxWidth: 600, mx: 'auto', p: 3 }}>
            <Paper sx={{ p: 4, textAlign: 'center' }}>
                <Typography variant="h5" gutterBottom>
                    Processing PDF
                </Typography>

                <Box sx={{ 
                    display: 'flex', 
                    flexDirection: 'column', 
                    alignItems: 'center',
                    my: 4 
                }}>
                    {status.status === 'processing' && (
                        <Box sx={{ position: 'relative', display: 'inline-flex', mb: 2 }}>
                            <CircularProgress size={80} />
                            <Box sx={{
                                position: 'absolute',
                                top: 0,
                                left: 0,
                                bottom: 0,
                                right: 0,
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                            }}>
                                <Typography variant="caption" component="div" color="text.secondary">
                                    {Math.round(status.progress)}%
                                </Typography>
                            </Box>
                        </Box>
                    )}

                    <Typography variant="body1" color="text.secondary" sx={{ mt: 2 }}>
                        {status.message}
                    </Typography>
                </Box>

                {status.status === 'completed' && (
                    <Alert severity="success" sx={{ mt: 2 }}>
                        Processing completed! Redirecting to card review...
                    </Alert>
                )}

                {status.status === 'failed' && (
                    <Alert severity="error" sx={{ mt: 2 }}>
                        Processing failed. Please try again.
                        {status.error && (
                            <Typography variant="body2" sx={{ mt: 1 }}>
                                Error: {status.error}
                            </Typography>
                        )}
                    </Alert>
                )}
            </Paper>
        </Box>
    );
}