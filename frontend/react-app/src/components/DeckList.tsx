import React, { useEffect, useState } from 'react';
import {
    Box,
    Typography,
    Paper,
    List,
    ListItem,
    ListItemText,
    Button,
    CircularProgress,
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import EditIcon from '@mui/icons-material/Edit';
import DownloadIcon from '@mui/icons-material/Download';

interface Deck {
    name: string;
    totalCards: number;
}

export default function DeckList() {
    const [decks, setDecks] = useState<Deck[]>([]);
    const [loading, setLoading] = useState(true);
    const navigate = useNavigate();

    useEffect(() => {
        fetchDecks();
    }, []);

    const fetchDecks = async () => {
        try {
            // Get list of CSV files from the cards directory
            const response = await fetch('/api/cards/list');
            const data = await response.json();
            // Ensure we only use the base name of each deck
            const processedDecks = data.decks.map((deck: { name: string; totalCards: number }) => ({
                ...deck,
                name: deck.name.split('/').pop() || deck.name // Get just the filename without path
            }));
            setDecks(processedDecks);
            setLoading(false);
        } catch (error) {
            console.error('Error fetching decks:', error);
            setLoading(false);
        }
    };

    const handleEdit = (deckName: string) => {
        console.log('Editing deck:', deckName); // Debug log
        // Get just the filename without path and extension
        const baseName = deckName.split('/').pop()?.replace('.csv', '') || deckName;
        console.log('Base name:', baseName); // Debug log
        const encodedDeckName = encodeURIComponent(baseName);
        console.log('Encoded deck name:', encodedDeckName); // Debug log
        const reviewPath = `/review/${encodedDeckName}`;
        console.log('Navigating to:', reviewPath); // Debug log
        navigate(reviewPath);
    };

    const handleDownload = async (deckName: string) => {
        try {
            // Get just the filename without path and extension
            const baseName = deckName.split('/').pop()?.replace('.csv', '') || deckName;
            const encodedDeckName = encodeURIComponent(baseName);
            console.log('Downloading deck:', baseName); // Debug log
            
            const response = await fetch(`/api/cards/apkg/${encodedDeckName}`, {
                method: 'GET',
            });
            
            if (!response.ok) {
                const errorText = await response.text();
                console.error('Failed to generate Anki deck:', response.status, errorText);
                throw new Error(`Failed to generate Anki deck: ${errorText}`);
            }

            // Check if we got an actual file
            const contentType = response.headers.get('content-type');
            if (!contentType || !contentType.includes('application/octet-stream')) {
                throw new Error('Invalid response format from server');
            }

            // Create a blob from the response
            const blob = await response.blob();
            if (blob.size === 0) {
                throw new Error('Generated file is empty');
            }
            
            // Create a download link and trigger it
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${baseName}.apkg`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
            
            console.log('Download completed successfully'); // Debug log
        } catch (error) {
            console.error('Error downloading deck:', error);
            alert('Failed to download deck: ' + (error instanceof Error ? error.message : 'Unknown error'));
        }
    };

    if (loading) {
        return (
            <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
                <CircularProgress />
            </Box>
        );
    }

    if (decks.length === 0) {
        return (
            <Box textAlign="center" p={4}>
                <Typography variant="h6">No decks available</Typography>
                <Typography variant="body1" color="textSecondary">
                    Upload some PDF files first to generate flashcards.
                </Typography>
            </Box>
        );
    }

    return (
        <Box sx={{ maxWidth: 800, mx: 'auto', p: 3 }}>
            <Typography variant="h4" gutterBottom>
                Available Decks
            </Typography>
            <Paper>
                <List>
                    {decks.map((deck) => (
                        <ListItem
                            key={deck.name}
                            divider
                            secondaryAction={
                                <Box>
                                    <Button
                                        startIcon={<EditIcon />}
                                        onClick={() => {
                                            console.log('Edit button clicked for deck:', deck.name); // Debug log
                                            handleEdit(deck.name);
                                        }}
                                        sx={{ mr: 1 }}
                                    >
                                        Edit
                                    </Button>
                                    <Button
                                        startIcon={<DownloadIcon />}
                                        onClick={() => handleDownload(deck.name)}
                                        color="primary"
                                    >
                                        Download
                                    </Button>
                                </Box>
                            }
                        >
                            <ListItemText
                                primary={deck.name}
                                secondary={`${deck.totalCards} cards`}
                            />
                        </ListItem>
                    ))}
                </List>
            </Paper>
        </Box>
    );
} 