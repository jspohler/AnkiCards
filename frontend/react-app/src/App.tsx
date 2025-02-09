import React from 'react';
import { Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import FileUpload from './components/FileUpload';
import ProcessingStatus from './components/ProcessingStatus';
import CardReview from './components/CardReview';
import DeckList from './components/DeckList';

function App() {
    return (
        <Layout>
            <Routes>
                <Route path="/" element={<FileUpload />} />
                <Route path="/status/:jobId" element={<ProcessingStatus />} />
                <Route path="/review/:deckName" element={<CardReview />} />
                <Route path="/decks" element={<DeckList />} />
            </Routes>
        </Layout>
    );
}

export default App; 