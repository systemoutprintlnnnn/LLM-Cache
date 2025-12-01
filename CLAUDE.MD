# CLAUDE.MD

## Project Overview

**LLM-Cache** is a high-performance semantic caching system for Large Language Models (LLMs) built with Golang. The primary goal is to significantly reduce LLM API costs by caching semantically similar queries, particularly beneficial for RAG-based (Retrieval-Augmented Generation) knowledge base systems.

## Project Purpose

- **Cost Reduction**: Minimize redundant LLM API calls by identifying semantically similar queries
- **Performance Optimization**: Fast cache lookups using semantic similarity matching
- **RAG Enhancement**: Specifically designed to optimize RAG-based knowledge systems

## Technology Stack

- **Language**: Go (Golang)
- **Focus**: Semantic caching, vector similarity search
- **Use Case**: LLM API optimization, RAG systems

## Key Concepts

### Semantic Caching
Unlike traditional key-value caching, semantic caching identifies queries with similar meaning:
- "What is the weather?" ≈ "Tell me about today's weather"
- Enables cache hits even when exact query text differs
- Uses embedding vectors and similarity metrics

### RAG Systems
This cache is optimized for Retrieval-Augmented Generation workflows:
- Frequently repeated knowledge base queries
- Similar user questions about the same topics
- Reduced latency and cost for common queries

## Development Guidelines

### Code Style
- Follow standard Go conventions and idioms
- Use `gofmt` for code formatting
- Maintain clear, descriptive variable and function names
- Add comments for complex algorithms, especially similarity matching logic

### Architecture Priorities
- **Performance**: Fast lookup and retrieval times
- **Scalability**: Handle large cache sizes efficiently
- **Accuracy**: High precision in semantic similarity matching
- **Memory Efficiency**: Optimize vector storage and comparison

### Testing
- Unit tests for core caching logic
- Benchmark tests for performance-critical paths
- Integration tests for end-to-end caching scenarios
- Test semantic similarity matching accuracy

## Working with This Codebase

### Common Tasks

1. **Adding New Cache Strategies**
   - Implement similarity scoring algorithms
   - Consider trade-offs between accuracy and speed

2. **Optimizing Performance**
   - Profile hot paths in similarity computation
   - Optimize vector operations and storage
   - Consider approximate nearest neighbor algorithms

3. **Extending Compatibility**
   - Support different embedding models
   - Add integrations with various LLM providers

### Important Considerations

- **Embedding Quality**: Cache effectiveness depends on embedding model quality
- **Similarity Threshold**: Balance between cache hit rate and accuracy
- **Cache Eviction**: Strategy for managing cache size and staleness
- **Concurrency**: Handle concurrent cache access safely

## Repository Structure

```
LLM-Cache/
├── LICENSE          # Project license
├── README.md        # Project description (Chinese)
└── CLAUDE.MD        # AI assistant context (this file)
```

## Future Development Areas

- [ ] Implement core semantic similarity engine
- [ ] Add support for multiple embedding providers
- [ ] Develop cache eviction strategies (LRU, LFU, semantic-based)
- [ ] Build monitoring and analytics dashboard
- [ ] Add distributed caching support
- [ ] Optimize for specific LLM providers (OpenAI, Anthropic, etc.)
- [ ] Implement cache warming strategies
- [ ] Add configuration management

## Notes for AI Assistants

When working on this project:
1. Prioritize performance - this is a caching system where speed matters
2. Consider the semantic similarity algorithms carefully
3. Think about cache hit rates and how to optimize them
4. Be mindful of memory usage with vector storage
5. Consider Chinese language support (project README is in Chinese)
6. Focus on production-ready, robust Go code
7. Document complex algorithms and design decisions

## Resources

- Go Best Practices: https://go.dev/doc/effective_go
- Vector Similarity: Consider algorithms like cosine similarity, dot product
- Embedding Models: OpenAI embeddings, sentence-transformers, etc.
- RAG Systems: Understanding retrieval patterns can inform cache design
