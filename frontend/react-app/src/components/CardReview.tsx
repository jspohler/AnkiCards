import React, { useState, useEffect } from 'react';
import {
  Box,
    Button,
  Card,
  CardContent,
    TextField,
  Typography,
  IconButton,
    CircularProgress,
} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import NavigateNextIcon from '@mui/icons-material/NavigateNext';
import NavigateBeforeIcon from '@mui/icons-material/NavigateBefore';
import { useParams } from 'react-router-dom';

interface FlashCard {
  question: string;
  answer: string;
}

export const CardReview: React.FC = () => {
    const { deckName } = useParams<{ deckName: string }>();
    const [cards, setCards] = useState<FlashCard[]>([]);
    const [currentIndex, setCurrentIndex] = useState(0);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [allCardsReviewed, setAllCardsReviewed] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        console.log('CardReview mounted with deckName:', deckName); // Debug log
        fetchCards();
    }, [deckName]);

    const fetchCards = async () => {
        if (!deckName) {
            console.error('No deck name provided');
            setError('No deck name provided');
            setLoading(false);
            return;
        }

        try {
            console.log('Fetching cards for deck:', deckName); // Debug log
            const encodedDeckName = encodeURIComponent(deckName);
            console.log('Encoded deck name:', encodedDeckName); // Debug log
            const response = await fetch(`/api/cards/csv/${encodedDeckName}`);
            
            if (!response.ok) {
                const errorText = await response.text();
                console.error('Failed to fetch cards:', response.status, errorText); // Debug log
                throw new Error(`Failed to fetch cards: ${response.statusText} (${errorText})`);
            }
            
            const data = await response.json();
            console.log('Received data:', data); // Debug log
            
            if (!data.cards || !Array.isArray(data.cards)) {
                throw new Error('Invalid response format');
            }
            
            setCards(data.cards);
            setError(null);
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
            console.error('Error fetching cards:', errorMessage);
            setError(errorMessage);
        } finally {
            setLoading(false);
        }
    };

    const handleUpdateCard = async (index: number, updatedCard: FlashCard) => {
        const newCards = [...cards];
        newCards[index] = updatedCard;
        setCards(newCards);
    };

    const handleDeleteCard = async (index: number) => {
        const newCards = cards.filter((_, i) => i !== index);
        setCards(newCards);
        if (currentIndex >= newCards.length) {
            setCurrentIndex(Math.max(0, newCards.length - 1));
        }
    };

    const handleSaveChanges = async () => {
        if (!deckName) {
            console.error('No deck name provided');
            return;
        }

        setSaving(true);
        try {
            const encodedDeckName = encodeURIComponent(deckName);
            const response = await fetch(`/api/cards/csv/${encodedDeckName}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(cards),
            });

            if (!response.ok) {
                throw new Error(`Failed to save cards: ${response.statusText}`);
            }

            // Show success feedback here if needed
        } catch (error) {
            console.error('Error saving cards:', error);
            // Show error feedback here if needed
        } finally {
            setSaving(false);
        }
    };

    const handleGenerateAnkiDeck = async () => {
        try {
            window.location.href = `/api/cards/apkg/${deckName}`;
        } catch (error) {
            console.error('Error generating Anki deck:', error);
        }
    };

    const handleNext = () => {
        if (currentIndex < cards.length - 1) {
            setCurrentIndex(currentIndex + 1);
        }
        if (currentIndex === cards.length - 2) {
            setAllCardsReviewed(true);
        }
    };

    const handlePrevious = () => {
        if (currentIndex > 0) {
            setCurrentIndex(currentIndex - 1);
        }
    };

    if (error) {
        return (
            <Box textAlign="center" p={4}>
                <Typography variant="h6" color="error" gutterBottom>
                    Error Loading Cards
                </Typography>
                <Typography color="error">{error}</Typography>
                <Button
                    variant="contained"
                    onClick={fetchCards}
                    sx={{ mt: 2 }}
                >
                    Retry
                </Button>
            </Box>
        );
    }

    if (loading) {
    return (
            <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
                <CircularProgress />
      </Box>
    );
  }

    if (cards.length === 0) {
    return (
            <Box textAlign="center" p={4}>
                <Typography variant="h6">No cards available</Typography>
      </Box>
    );
  }

    const currentCard = cards[currentIndex];

  return (
        <Box sx={{ maxWidth: 800, mx: 'auto', p: 3 }}>
            <Typography variant="h5" gutterBottom textAlign="center">
                Card Review ({currentIndex + 1} / {cards.length})
      </Typography>

            <Card sx={{ mb: 3, minHeight: 200 }}>
                  <CardContent>
                    <Box mb={2}>
                        <Typography variant="subtitle1" color="textSecondary" gutterBottom>
                            Question:
                        </Typography>
                        <TextField
                            fullWidth
                            multiline
                            rows={2}
                            value={currentCard.question}
                            onChange={(e) =>
                                handleUpdateCard(currentIndex, {
                                    ...currentCard,
                                    question: e.target.value,
                                })
                            }
                        />
                    </Box>

                    <Box>
                        <Typography variant="subtitle1" color="textSecondary" gutterBottom>
                            Answer:
                        </Typography>
                        <TextField
                            fullWidth
                            multiline
                            rows={2}
                            value={currentCard.answer}
                            onChange={(e) =>
                                handleUpdateCard(currentIndex, {
                                    ...currentCard,
                                    answer: e.target.value,
                                })
                            }
                        />
                    </Box>
                </CardContent>
            </Card>

            <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
                <IconButton onClick={handlePrevious} disabled={currentIndex === 0}>
                    <NavigateBeforeIcon />
                </IconButton>

                <Box>
                      <IconButton
                        onClick={() => handleDeleteCard(currentIndex)}
                        color="error"
                        sx={{ mr: 1 }}
                      >
                        <DeleteIcon />
                      </IconButton>
                    <Button
                        variant="contained"
                        onClick={handleSaveChanges}
                        disabled={saving}
                        sx={{ mr: 1 }}
                    >
                        {saving ? <CircularProgress size={24} /> : 'Save Changes'}
                    </Button>
                    {allCardsReviewed && (
                        <Button
                            variant="contained"
                            color="primary"
                            onClick={handleGenerateAnkiDeck}
                        >
                            Generate Anki Deck
                        </Button>
                    )}
                    </Box>

                <IconButton
                    onClick={handleNext}
                    disabled={currentIndex === cards.length - 1}
                >
                    <NavigateNextIcon />
                </IconButton>
            </Box>

            <Box textAlign="center">
                <Typography variant="body2" color="textSecondary">
                    {allCardsReviewed
                        ? 'All cards reviewed! You can now generate the Anki deck.'
                        : 'Review all cards to enable Anki deck generation.'}
          </Typography>
            </Box>
    </Box>
  );
};

export default CardReview; 